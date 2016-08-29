// Copyright 2013 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
)

// String constants for instrumentation.
const (
	namespace = "prometheus"
	subsystem = "remote_storage"

	result  = "result"
	success = "success"
	failure = "failure"
	dropped = "dropped"
)

// StorageClient defines an interface for sending a batch of samples to an
// external timeseries database.
type StorageClient interface {
	// Store stores the given samples in the remote storage.
	Store(model.Samples) error
	// Name identifies the remote storage implementation.
	Name() string
}

type StorageQueueManagerConfig struct {
	QueueCapacity     int           // Number of samples to buffer per shard before we start dropping them.
	Shards            int           // Number of shards, i.e. amount of concurrency.
	MaxSamplesPerSend int           // Maximum number of samples per send.
	BatchSendDeadline time.Duration // Maximum time sample will wait in buffer.
}

var defaultConfig = StorageQueueManagerConfig{
	QueueCapacity:     100 * 1024 / 10,
	Shards:            10,
	MaxSamplesPerSend: 100,
	BatchSendDeadline: 5 * time.Second,
}

// StorageQueueManager manages a queue of samples to be sent to the Storage
// indicated by the provided StorageClient.
type StorageQueueManager struct {
	cfg    StorageQueueManagerConfig
	tsdb   StorageClient
	shards []chan *model.Sample
	wg     sync.WaitGroup
	done   chan struct{}

	samplesCount  *prometheus.CounterVec
	sendLatency   prometheus.Summary
	failedBatches prometheus.Counter
	failedSamples prometheus.Counter
	queueLength   prometheus.Gauge
	queueCapacity prometheus.Metric
}

// NewStorageQueueManager builds a new StorageQueueManager.
func NewStorageQueueManager(tsdb StorageClient, cfg StorageQueueManagerConfig) *StorageQueueManager {
	constLabels := prometheus.Labels{
		"type": tsdb.Name(),
	}

	shards := make([]chan *model.Sample, cfg.Shards)
	for i := 0; i < cfg.Shards; i++ {
		shards[i] = make(chan *model.Sample, cfg.QueueCapacity)
	}

	t := &StorageQueueManager{
		cfg:    cfg,
		tsdb:   tsdb,
		shards: shards,
		done:   make(chan struct{}),

		samplesCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "sent_samples_total",
				Help:        "Total number of processed samples to be sent to remote storage.",
				ConstLabels: constLabels,
			},
			[]string{result},
		),
		sendLatency: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "send_latency_seconds",
			Help:        "Latency quantiles for sending sample batches to the remote storage.",
			ConstLabels: constLabels,
		}),
		failedBatches: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "failed_batches_total",
			Help:        "Total number of sample batches that encountered an error while being sent to the remote storage.",
			ConstLabels: constLabels,
		}),
		failedSamples: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "failed_samples_total",
			Help:        "Total number of samples that encountered an error while being sent to the remote storage.",
			ConstLabels: constLabels,
		}),
		queueLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "queue_length",
			Help:        "The number of processed samples queued to be sent to the remote storage.",
			ConstLabels: constLabels,
		}),
		queueCapacity: prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "queue_capacity"),
				"The capacity of the queue of samples to be sent to the remote storage.",
				nil,
				constLabels,
			),
			prometheus.GaugeValue,
			float64(cfg.QueueCapacity),
		),
	}

	t.wg.Add(cfg.Shards)
	return t
}

// Append queues a sample to be sent to the remote storage. It drops the
// sample on the floor if the queue is full.
// Always returns nil.
func (t *StorageQueueManager) Append(s *model.Sample) error {
	fp := s.Metric.FastFingerprint()
	shard := uint64(fp) % uint64(t.cfg.Shards)

	select {
	case t.shards[shard] <- s:
	default:
		t.samplesCount.WithLabelValues(dropped).Inc()
		log.Warn("Remote storage queue full, discarding sample.")
	}
	return nil
}

// NeedsThrottling implements storage.SampleAppender. It will always return
// false as a remote storage drops samples on the floor if backlogging instead
// of asking for throttling.
func (*StorageQueueManager) NeedsThrottling() bool {
	return false
}

// Describe implements prometheus.Collector.
func (t *StorageQueueManager) Describe(ch chan<- *prometheus.Desc) {
	t.samplesCount.Describe(ch)
	t.sendLatency.Describe(ch)
	ch <- t.failedBatches.Desc()
	ch <- t.failedSamples.Desc()
	ch <- t.queueLength.Desc()
	ch <- t.queueCapacity.Desc()
}

// QueueLength returns the number of outstanding samples in the queue.
func (t *StorageQueueManager) QueueLength() int {
	queueLength := 0
	for _, shard := range t.shards {
		queueLength += len(shard)
	}
	return queueLength
}

// Collect implements prometheus.Collector.
func (t *StorageQueueManager) Collect(ch chan<- prometheus.Metric) {
	t.samplesCount.Collect(ch)
	t.sendLatency.Collect(ch)
	t.queueLength.Set(float64(t.QueueLength()))
	ch <- t.failedBatches
	ch <- t.failedSamples
	ch <- t.queueLength
	ch <- t.queueCapacity
}

// Run continuously sends samples to the remote storage.
func (t *StorageQueueManager) Run() {
	for i := 0; i < t.cfg.Shards; i++ {
		go t.runShard(i)
	}
	t.wg.Wait()
}

// Stop stops sending samples to the remote storage and waits for pending
// sends to complete.
func (t *StorageQueueManager) Stop() {
	log.Infof("Stopping remote storage...")
	for _, shard := range t.shards {
		close(shard)
	}
	t.wg.Wait()
	log.Info("Remote storage stopped.")
}

func (t *StorageQueueManager) runShard(i int) {
	defer t.wg.Done()
	shard := t.shards[i]

	// Send batches of at most MaxSamplesPerSend samples to the remote storage.
	// If we have fewer samples than that, flush them out after a deadline
	// anyways.
	pendingSamples := model.Samples{}

	for {
		select {
		case s, ok := <-shard:
			if !ok {
				if len(pendingSamples) > 0 {
					log.Infof("Flushing %d samples to remote storage...", len(pendingSamples))
					t.sendSamples(pendingSamples)
					log.Infof("Done flushing.")
				}
				return
			}

			pendingSamples = append(pendingSamples, s)

			for len(pendingSamples) >= t.cfg.MaxSamplesPerSend {
				t.sendSamples(pendingSamples[:t.cfg.MaxSamplesPerSend])
				pendingSamples = pendingSamples[t.cfg.MaxSamplesPerSend:]
			}
		case <-time.After(t.cfg.BatchSendDeadline):
			if len(pendingSamples) > 0 {
				t.sendSamples(pendingSamples)
				pendingSamples = pendingSamples[:0]
			}
		}
	}
}

func (t *StorageQueueManager) sendSamples(s model.Samples) {
	// Samples are sent to the remote storage on a best-effort basis. If a
	// sample isn't sent correctly the first time, it's simply dropped on the
	// floor.
	begin := time.Now()
	err := t.tsdb.Store(s)
	duration := time.Since(begin).Seconds()

	labelValue := success
	if err != nil {
		log.Warnf("error sending %d samples to remote storage: %s", len(s), err)
		labelValue = failure
		t.failedBatches.Inc()
		t.failedSamples.Add(float64(len(s)))
	}
	t.samplesCount.WithLabelValues(labelValue).Add(float64(len(s)))
	t.sendLatency.Observe(duration)
}

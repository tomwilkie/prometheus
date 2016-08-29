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
	"fmt"
	"sync"
	"testing"

	"github.com/prometheus/common/model"
)

type TestStorageClient struct {
	receivedSamples map[string]model.Samples
	expectedSamples map[string]model.Samples
	wg              sync.WaitGroup
	mtx             sync.Mutex
}

func NewTestStorageClient() *TestStorageClient {
	return &TestStorageClient{
		receivedSamples: map[string]model.Samples{},
		expectedSamples: map[string]model.Samples{},
	}
}

func (c *TestStorageClient) expectSamples(ss model.Samples) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, s := range ss {
		ts := s.Metric.String()
		c.expectedSamples[ts] = append(c.expectedSamples[ts], s)
	}
	c.wg.Add(len(ss))
}

func (c *TestStorageClient) waitForExpectedSamples(t *testing.T) {
	c.wg.Wait()

	c.mtx.Lock()
	defer c.mtx.Unlock()
	for ts, expectedSamples := range c.expectedSamples {
		for i, expected := range expectedSamples {
			if !expected.Equal(c.receivedSamples[ts][i]) {
				t.Fatalf("%d. Expected %v, got %v", i, expected, c.receivedSamples[ts][i])
			}
		}
	}
}

func (c *TestStorageClient) Store(ss model.Samples) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, s := range ss {
		ts := s.Metric.String()
		c.receivedSamples[ts] = append(c.receivedSamples[ts], s)
	}
	c.wg.Add(-len(ss))
	return nil
}

func (c *TestStorageClient) Name() string {
	return "teststorageclient"
}

type TestBlockingStorageClient struct {
	block   chan bool
	getData chan bool
}

func NewTestBlockedStorageClient() *TestBlockingStorageClient {
	return &TestBlockingStorageClient{
		block:   make(chan bool),
		getData: make(chan bool),
	}
}

func (c *TestBlockingStorageClient) Store(s model.Samples) error {
	<-c.getData
	<-c.block
	return nil
}

func (c *TestBlockingStorageClient) unlock() {
	close(c.getData)
	close(c.block)
}

func (c *TestBlockingStorageClient) Name() string {
	return "testblockingstorageclient"
}

func TestSampleDelivery(t *testing.T) {
	// Let's create an even number of send batches so we don't run into the
	// batch timeout case.
	cfg := defaultConfig
	n := cfg.QueueCapacity * 2
	cfg.Shards = 1

	samples := make(model.Samples, 0, n)
	for i := 0; i < n; i++ {
		name := model.LabelValue(fmt.Sprintf("test_metric_%d", i))
		samples = append(samples, &model.Sample{
			Metric: model.Metric{
				model.MetricNameLabel: name,
			},
			Value: model.SampleValue(i),
		})
	}

	c := NewTestStorageClient()
	c.expectSamples(samples[:len(samples)/2])
	m := NewStorageQueueManager(c, cfg)

	// These should be received by the client.
	for _, s := range samples[:len(samples)/2] {
		m.Append(s)
	}
	// These will be dropped because the queue is full.
	for _, s := range samples[len(samples)/2:] {
		m.Append(s)
	}
	go m.Run()
	defer m.Stop()

	c.waitForExpectedSamples(t)
}

func TestSampleDeliveryOrder(t *testing.T) {
	cfg := defaultConfig
	ts := 10
	n := cfg.MaxSamplesPerSend * ts
	// Ensure we don't drop samples in this test.
	cfg.QueueCapacity = n

	samples := make(model.Samples, 0, n)
	for i := 0; i < n; i++ {
		name := model.LabelValue(fmt.Sprintf("test_metric_%d", i%ts))
		samples = append(samples, &model.Sample{
			Metric: model.Metric{
				model.MetricNameLabel: name,
			},
			Value:     model.SampleValue(i),
			Timestamp: model.Time(i),
		})
	}

	c := NewTestStorageClient()
	c.expectSamples(samples)
	m := NewStorageQueueManager(c, cfg)

	// These should be received by the client.
	for _, s := range samples {
		m.Append(s)
	}
	go m.Run()
	defer m.Stop()

	c.waitForExpectedSamples(t)
}

func TestSpawnNotMoreThanMaxConcurrentSendsGoroutines(t *testing.T) {
	// `maxSamplesPerSend*maxConcurrentSends` samples should be consumed by
	//  goroutines, `maxSamplesPerSend` should be still in the queue.
	cfg := defaultConfig
	n := cfg.MaxSamplesPerSend*cfg.Shards + cfg.MaxSamplesPerSend
	cfg.QueueCapacity = n

	samples := make(model.Samples, 0, n)
	for i := 0; i < n; i++ {
		name := model.LabelValue(fmt.Sprintf("test_metric_%d", i))
		samples = append(samples, &model.Sample{
			Metric: model.Metric{
				model.MetricNameLabel: name,
			},
			Value: model.SampleValue(i),
		})
	}

	c := NewTestBlockedStorageClient()
	m := NewStorageQueueManager(c, cfg)

	go m.Run()

	for _, s := range samples {
		m.Append(s)
	}

	for i := 0; i < cfg.Shards; i++ {
		c.getData <- true // Wait while all goroutines are spawned.
	}

	queueLength := m.QueueLength()
	if queueLength != cfg.MaxSamplesPerSend {
		t.Errorf("Queue should contain %d samples, it contains %d.", cfg.MaxSamplesPerSend, queueLength)
	}

	c.unlock()

	defer m.Stop()
}

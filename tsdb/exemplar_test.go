// Copyright 2020 The Prometheus Authors
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

package tsdb

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/prometheus/pkg/exemplar"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

func TestAddExemplar(t *testing.T) {
	exs, err := NewCircularExemplarStorage(2, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l := labels.Labels{
		{Name: "service", Value: "asdf"},
	}
	e := exemplar.Exemplar{
		Labels: labels.Labels{
			labels.Label{
				Name:  "traceID",
				Value: "qwerty",
			},
		},
		Value: 0.1,
		Ts:    1,
	}

	err = es.AddExemplar(l, e)
	require.NoError(t, err)
	require.Equal(t, es.index[l.String()].last, 0, "exemplar was not stored correctly")

	e2 := exemplar.Exemplar{
		Labels: labels.Labels{
			labels.Label{
				Name:  "traceID",
				Value: "zxcvb",
			},
		},
		Value: 0.1,
		Ts:    2,
	}

	err = es.AddExemplar(l, e2)
	require.NoError(t, err)
	require.Equal(t, es.index[l.String()].last, 1, "exemplar was not stored correctly, location of newest exemplar for series in index did not update")
	require.True(t, es.exemplars[es.index[l.String()].last].exemplar.Equals(e2), "exemplar was not stored correctly, expected %+v got: %+v", e2, es.exemplars[es.index[l.String()].last].exemplar)

	err = es.AddExemplar(l, e2)
	require.NoError(t, err, "no error is expected attempting to add duplicate exemplar")

	e3 := e2
	e3.Ts = 3
	err = es.AddExemplar(l, e3)
	require.NoError(t, err, "no error is expected when attempting to add duplicate exemplar, even with different timestamp")

	e3.Ts = 1
	e3.Value = 0.3
	err = es.AddExemplar(l, e3)
	require.Equal(t, err, storage.ErrOutOfOrderExemplar)
}

func TestAddExtraExemplar(t *testing.T) {
	// Test that circular buffer index and assignment
	// works properly, adding more exemplars than can
	// be stored and then querying for them.
	exs, err := NewCircularExemplarStorage(5, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l := labels.Labels{
		{Name: "service", Value: "asdf"},
	}

	var eList []exemplar.Exemplar
	for i := 0; i < len(es.exemplars)+1; i++ {
		e := exemplar.Exemplar{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "a",
				},
			},
			Value: float64(i+1) / 10,
			Ts:    int64(101 + i),
		}
		es.AddExemplar(l, e)
		eList = append(eList, e)
	}
	require.True(t, (es.exemplars[0].exemplar.Ts == 106), "exemplar was not stored correctly")

	ret, err := es.Select(100, 110, l)
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(eList[1:], ret), "select did not return expected exemplars\n\texpected: %+v\n\tactual: %+v\n", eList[1:], ret)
}

func TestSelectExemplar(t *testing.T) {
	exs, err := NewCircularExemplarStorage(5, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l := labels.Labels{
		{Name: "service", Value: "asdf"},
	}
	e := exemplar.Exemplar{
		Labels: labels.Labels{
			labels.Label{
				Name:  "traceID",
				Value: "qwerty",
			},
		},
		Value: 0.1,
		Ts:    12,
	}

	es.AddExemplar(l, e)
	require.True(t, reflect.DeepEqual(es.exemplars[0].exemplar, e), "exemplar was not stored correctly")

	exemplars, err := es.Select(0, 100, l)
	require.NoError(t, err)

	expectedResult := []exemplar.Exemplar{e}
	require.True(t, reflect.DeepEqual(expectedResult, exemplars), "select did not return expected exemplars\n\texpected: %+v\n\tactual: %+v\n", expectedResult, exemplars)
}

func TestSelectExemplar_MultiSeries(t *testing.T) {
	exs, err := NewCircularExemplarStorage(5, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l1 := labels.Labels{
		{Name: "__name__", Value: "test_metric"},
		{Name: "service", Value: "asdf"},
	}
	l2 := labels.Labels{
		{Name: "__name__", Value: "test_metric2"},
		{Name: "service", Value: "qwer"},
	}

	for i := 0; i < len(es.exemplars); i++ {
		e1 := exemplar.Exemplar{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "a",
				},
			},
			Value: float64(i+1) / 10,
			Ts:    int64(101 + i),
		}
		err = es.AddExemplar(l1, e1)
		require.NoError(t, err)

		e2 := exemplar.Exemplar{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "b",
				},
			},
			Value: float64(i+1) / 10,
			Ts:    int64(101 + i),
		}
		err = es.AddExemplar(l2, e2)
		require.NoError(t, err)
	}

	el, err := es.Select(100, 200, l2)
	require.NoError(t, err)
	require.True(t, len(el) == 3, "didn't get expected 8 exemplars, got %d", len(el))

	el, err = es.Select(100, 200, l1)
	require.NoError(t, err)
	require.True(t, len(el) == 2, "didn't get expected 8 exemplars, got %d", len(el))
}

func TestSelectExemplar_TimeRange(t *testing.T) {
	exs, err := NewCircularExemplarStorage(4, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l := labels.Labels{
		{Name: "service", Value: "asdf"},
	}
	exemplars := []exemplar.Exemplar{
		{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "qwerty",
				},
			},
			Value: 0.1,
			Ts:    101,
		},
		{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "zxcvbn",
				},
			},
			Value: 0.1,
			Ts:    102,
		},
		{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "asdfgh",
				},
			},
			Value: 0.1,
			Ts:    103,
		},
		{
			Labels: labels.Labels{
				labels.Label{
					Name:  "traceID",
					Value: "hjkl;",
				},
			},
			Value: 0.1,
			Ts:    106,
		},
	}

	for i, e := range exemplars {
		err := es.AddExemplar(l, e)
		require.NoError(t, err)
		require.Equal(t, es.index[l.String()].last, i, "exemplar was not stored correctly")
	}

	el, err := es.Select(102, 105, l)
	require.NoError(t, err)
	require.True(t, len(el) == 2, "didn't get expected one exemplar")
	require.True(t, reflect.DeepEqual(el, exemplars[1:3]), "returned exemplar did not matched expected\n\tactual: %+v\n\texpected %+v", el, exemplars[1:3])
}

func TestIndexOverwrite(t *testing.T) {
	exs, err := NewCircularExemplarStorage(2, nil)
	require.NoError(t, err)
	es := exs.(*CircularExemplarStorage)

	l1 := labels.Labels{
		{Name: "service", Value: "asdf"},
	}

	l2 := labels.Labels{
		{Name: "service", Value: "qwer"},
	}

	err = es.AddExemplar(l1, exemplar.Exemplar{Value: 1, Ts: 1})
	require.NoError(t, err)
	err = es.AddExemplar(l2, exemplar.Exemplar{Value: 2, Ts: 2})
	require.NoError(t, err)
	err = es.AddExemplar(l2, exemplar.Exemplar{Value: 3, Ts: 3})
	require.NoError(t, err)

	_, ok := es.index[l1.String()]
	require.False(t, ok)
	require.Equal(t, &indexEntry{1, 0}, es.index[l2.String()])

	err = es.AddExemplar(l1, exemplar.Exemplar{Value: 4, Ts: 4})
	require.NoError(t, err)

	i := es.index[l2.String()]
	require.Equal(t, &indexEntry{0, 0}, i)
}

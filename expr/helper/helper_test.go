package helper

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestSummarizeValues(t *testing.T) {
	epsilon := math.Nextafter(1, 2) - 1
	tests := []struct {
		name     string
		function string
		values   []float64
		expected float64
	}{
		{
			name:     "no values",
			function: "sum",
			values:   []float64{},
			expected: math.NaN(),
		},
		{
			name:     "sum",
			function: "sum",
			values:   []float64{1, 2, 3},
			expected: 6,
		},
		{
			name:     "sum alias",
			function: "total",
			values:   []float64{1, 2, 3},
			expected: 6,
		},
		{
			name:     "avg",
			function: "avg",
			values:   []float64{1, 2, 3, 4},
			expected: 2.5,
		},
		{
			name:     "max",
			function: "max",
			values:   []float64{1, 2, 3, 4},
			expected: 4,
		},
		{
			name:     "min",
			function: "min",
			values:   []float64{1, 2, 3, 4},
			expected: 1,
		},
		{
			name:     "last",
			function: "last",
			values:   []float64{1, 2, 3, 4},
			expected: 4,
		},
		{
			name:     "range",
			function: "range",
			values:   []float64{1, 2, 3, 4},
			expected: 3,
		},
		{
			name:     "median",
			function: "median",
			values:   []float64{1, 2, 3, 10, 11},
			expected: 3,
		},
		{
			name:     "multiply",
			function: "multiply",
			values:   []float64{1, 2, 3, 4},
			expected: 24,
		},
		{
			name:     "diff",
			function: "diff",
			values:   []float64{1, 2, 3, 4},
			expected: -8,
		},
		{
			name:     "count",
			function: "count",
			values:   []float64{1, 2, 3, 4},
			expected: 4,
		},
		{
			name:     "stddev",
			function: "stddev",
			values:   []float64{1, 2, 3, 4},
			expected: 1.118033988749895,
		},
		{
			name:     "p50 (fallback)",
			function: "p50",
			values:   []float64{1, 2, 3, 10, 11},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SummarizeValues(tt.function, tt.values)
			if math.Abs(actual-tt.expected) > epsilon {
				t.Errorf("actual %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	testCases := []struct {
		metric string
		result map[string]string
	}{
		{
			"cpu.usage_idle;cpu=cpu-total;host=test",
			map[string]string{"name": "cpu.usage_idle", "host": "test", "cpu": "cpu-total"},
		},
		{
			"x.y.z;foo1=foo2;=bar1;baz1=",
			map[string]string{"name": "x.y.z", "foo1": "foo2", "baz1": "", "": "bar1"},
		},
		{
			"x.y.z",
			map[string]string{"name": "x.y.z"},
		},
		{
			"func1(func2(x.y.z;key1=value1;key2=value2))",
			map[string]string{"name": "x.y.z", "key1": "value1", "key2": "value2"},
		},
		{
			"func1(func2(func3(x.y.z;key1=value1, param31), param21, param22))",
			map[string]string{"name": "x.y.z", "key1": "value1"},
		},
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("Checking %s", tt.metric), func(t *testing.T) {
			actual := ExtractTags(tt.metric)
			if !reflect.DeepEqual(actual, tt.result) {
				t.Errorf("actual %v, expected %v", actual, tt.result)
			}
		})
	}
}

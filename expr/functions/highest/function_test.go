package highest

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"math"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestHighestMultiReturn(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			parser.NewExpr("highestCurrent",
				"metric1",
				2,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			"highestCurrent",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32)},
			},
		},
		{
			parser.NewExpr("highestCurrent",
				"metric1",
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			"highestCurrent",
			map[string][]*types.MetricData{
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32)},
			},
		},

		{
			parser.NewExpr("highestMax",
				"metric1", 2,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12, 9}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 5, 5, 5, 5, 5, 3}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10, 10}, 1, now32),
				},
			},
			"highestMax",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12, 9}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10, 10}, 1, now32)},
			},
		},
		{
			parser.NewExpr("highestMin",
				"metric1", 2,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{6, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 2, 3, 3, 4, 10}, 1, now32),
				},
			},
			"highestMin",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{3, 2, 3, 3, 4, 10}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}
}

func TestHighest(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("highestCurrent",
				"metric1",
				1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricC", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 1, 3, 3, 4, 15}, 1, now32)},
			nil,
		},
		{
			parser.NewExpr("highestCurrent",
				"metric1",
				4,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
				//NOTE(nnuss): highest* functions filter null-valued series as a side-effect when `n` >= number of series
				//TODO(nnuss): bring lowest* functions into harmony with this side effect or get rid of it
				//types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
			},
			nil,
		},
		{
			parser.NewExpr("highestAverage",
				"metric1",
				1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 5, 5, 5, 5, 5}, 1, now32)},
			nil,
		},

		{
			parser.NewExpr("highestMax",
				"metric1",
				1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12, 9}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 5, 5, 5, 5, 5, 3}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA",
				[]float64{1, 1, 3, 3, 4, 12, 9}, 1, now32)},
			nil,
		},
		{
			parser.NewExpr("highestMin",
				"metric1",
				1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{6, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB",
				[]float64{2, 5, 5, 5, 5, 5}, 1, now32)},
			nil,
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt, true)
		})
	}
}

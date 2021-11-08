package compareWithThreshold

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
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

func TestCompareWithThreshold_Tagged(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("compareWithThreshold",
				"seriesByTag('name=series')",
				"seriesByTag('name=series_threshold')",
				10,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('name=series')", 0, 1}: {
					types.MakeMetricData("series;type=horse", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series;type=chicken", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),

					// these two metrics don't have a threshold series and use the default value
					types.MakeMetricData("series;type=noThreshold_1", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series;type=noThreshold_2", []float64{1, math.NaN(), 3, 12, 4, math.NaN()}, 1, now32),
				},
				{"seriesByTag('name=series_threshold')", 0, 1}: {
					types.MakeMetricData("series_threshold;type=horse", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("series_threshold;type=chicken", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("series;type=chicken", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
				types.MakeMetricData("series;type=noThreshold_2", []float64{1, math.NaN(), 3, 12, 4, math.NaN()}, 1, now32),
			},
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

func TestCompareWithThreshold_Untagged(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("compareWithThreshold",
				"series.*.*",
				"__thresholds",
				10,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"series.*.*", 0, 1}: {
					types.MakeMetricData("series.one.horse", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series.one.chicken", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),

					// these two metrics don't have a threshold series and use the default value
					types.MakeMetricData("series.no_threshold.1", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series.no_threshold.2", []float64{1, math.NaN(), 3, 12, 4, math.NaN()}, 1, now32),
				},
				{"__thresholds", 0, 1}: {
					types.MakeMetricData("series.one.horse", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("series.one.chicken", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("series.one.chicken", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
				types.MakeMetricData("series.no_threshold.2", []float64{1, math.NaN(), 3, 12, 4, math.NaN()}, 1, now32),
			},
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

package threshold

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

func TestThreshold_Tagged(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("threshold",
				"seriesByTag('name=series')",
				"seriesByTag('name=series_threshold')",
				10,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('name=series')", 0, 1}: {
					types.MakeMetricData("series;type=horse", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series;type=chicken", []float64{1, math.NaN(), 6, 3, 7, math.NaN()}, 1, now32),

					// these two metrics don't have a threshold series and use the default value
					types.MakeMetricData("series;type=noThreshold_1", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series;type=noThreshold_2", []float64{1, math.NaN(), 11, 12, 4, math.NaN()}, 1, now32),
				},
				{"seriesByTag('name=series_threshold')", 0, 1}: {
					types.MakeMetricData("series_threshold;type=horse", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("series_threshold;type=chicken", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("series;type=chicken", []float64{1, math.NaN(), 6, 3, 7, math.NaN()}, 1, now32),
				types.MakeMetricData("series;type=noThreshold_2", []float64{1, math.NaN(), 11, 12, 4, math.NaN()}, 1, now32),
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

func TestThreshold_Untagged(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("threshold",
				"series.*.*",
				"__thresholds",
				10,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"series.*.*", 0, 1}: {
					types.MakeMetricData("series.one.horse", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series.one.chicken", []float64{1, math.NaN(), 6, 3, 7, math.NaN()}, 1, now32),

					// these two metrics don't have a threshold series and use the default value
					types.MakeMetricData("series.no_threshold.1", []float64{1, math.NaN(), 3, 5, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("series.no_threshold.2", []float64{1, math.NaN(), 11, 12, 4, math.NaN()}, 1, now32),
				},
				{"__thresholds", 0, 1}: {
					types.MakeMetricData("series.one.horse", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("series.one.chicken", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("series.one.chicken", []float64{1, math.NaN(), 6, 3, 7, math.NaN()}, 1, now32),
				types.MakeMetricData("series.no_threshold.2", []float64{1, math.NaN(), 11, 12, 4, math.NaN()}, 1, now32),
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

func TestThreshold_MissingDataInThresholds(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("threshold",
				"series.*.*",
				"__thresholds",
				10,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"series.*.*", 0, 1}: {
					types.MakeMetricData("series.one.small_gap_return", []float64{11, math.NaN(), 3, 15, 14, math.NaN()}, 1, now32),
					types.MakeMetricData("series.one.small_gap_skip", []float64{
						15, 15,
						15, 15, 15, 15, 15, 15, 15, 15, 15, 15, // 10 points here
						15,
					}, 1, now32),
					types.MakeMetricData("series.one.large_gap", []float64{
						7, 7,
						7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 12 points here
						7,
					}, 1, now32),
					types.MakeMetricData("series.one.gap_in_front", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("series.one.gap_in_front_2", []float64{15, 15, 15, 15, 15, 15}, 1, now32),
				},
				{"__thresholds", 0, 1}: {
					types.MakeMetricData("series.one.small_gap_return", []float64{5, math.NaN(), math.NaN(), math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("series.one.small_gap_skip", []float64{
						20, 20,

						math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
						math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), // 10 points here

						20,
					}, 1, now32),
					types.MakeMetricData("series.one.large_gap", []float64{
						5, 5,

						math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
						math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), // 12 points here

						5,
					}, 1, now32),
					types.MakeMetricData("series.one.gap_in_front", []float64{math.NaN(), math.NaN(), 5, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("series.one.gap_in_front_2", []float64{math.NaN(), math.NaN(), 17, math.NaN(), 17, 17}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("series.one.small_gap_return", []float64{11, math.NaN(), 3, 15, 14, math.NaN()}, 1, now32),
				types.MakeMetricData("series.one.large_gap", []float64{
					7, 7,

					7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 12 points

					7,
				}, 1, now32),
				types.MakeMetricData("series.one.gap_in_front", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
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

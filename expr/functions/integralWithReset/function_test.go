package integralWithReset

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

func TestIntegralWithResetMultiReturn(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			parser.NewExpr("integralWithReset",

				"metric[12]",
				"reset",
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 1, 3, 5, 8, 13, 21}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 1, 1, 1, 1, 1, 1}, 1, now32),
				},
				{"reset", 0, 1}: {
					types.MakeMetricData("reset", []float64{0, 0, 0, 1, 1, 0, 0}, 1, now32),
				},
			},
			"integralWithReset",
			map[string][]*types.MetricData{
				"integralWithReset(metric1,reset)": {types.MakeMetricData(
					"integralWithReset(metric1,reset)",
					[]float64{1, 2, 5, 0, 0, 13, 34},
					1, now32,
				)},
				"integralWithReset(metric2,reset)": {types.MakeMetricData(
					"integralWithReset(metric2,reset)",
					[]float64{1, 2, 3, 0, 0, 1, 2},
					1, now32,
				)},
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

func TestIntegralWithReset(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("integralWithReset",
				"metric1", "metric2",
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12, 15}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0, math.NaN(), 0, math.NaN(), 0, 6, 0}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("integralWithReset(metric1,metric2)",
				[]float64{1, math.NaN(), math.NaN(), 4, 8, 0, 15}, 1, now32)},
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

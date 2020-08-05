package slo

import (
	"math"
	"testing"
	"time"

	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
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

func TestSlo_Do(t *testing.T) {
	nan := math.NaN()
	now32 := int32(time.Now().Unix())

	testCases := []th.EvalTestItem{
		{
			parser.NewExpr("slo", parser.ArgName("x.y.z"), parser.ArgValue("10sec"), parser.ArgValue("above"), 2),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, 2, 3, 4, 5, nan, nan, 6, 7, 0, 8},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"slo(x.y.z, 10sec, above, 2)",
					// (1, 2) -> 0
					// (3, 4) -> 1
					// (5, nan) -> 1: all not-null elements are above 2
					// (nan, 6) -> 1: the same
					// (7, 0) -> 0.5: only 1 element of 2 is above 2
					// (8) -> 1
					[]float64{0, 1, 1, 1, 0.5, 1},
					10,
					now32,
				),
			},
			nil,
		},
		{
			parser.NewExpr("slo", parser.ArgName("x.y.z"), parser.ArgValue("4sec"), parser.ArgValue("below"), 6),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"slo(x.y.z, 4sec, below, 6)",
					// all data points are nan because interval (4 sec) is less than step time (5 sec)
					[]float64{nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan},
					4,
					now32,
				),
			},
			nil,
		},
	}

	for _, testCase := range testCases {
		testName := testCase.E.Target() + "(" + testCase.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase, false)
		})
	}
}

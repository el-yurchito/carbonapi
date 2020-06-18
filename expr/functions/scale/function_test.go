package scale

import (
	"fmt"
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

func TestScale_Do(t *testing.T) {
	nan := math.NaN()
	now32 := int32(time.Now().Unix())

	testCases := []th.EvalTestItem{
		{
			parser.NewExpr("scale", parser.ArgName("x.y.z"), 1.5),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, 2, 3, 4, nan, 0, nan, 5, 6},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"scale(x.y.z,1.5)",
					[]float64{1.5, 3, 4.5, 6, nan, 0, nan, 7.5, 9},
					5,
					now32,
				),
			},
			nil,
		},
		{
			parser.NewExpr("scaleAfterTimestamp", parser.ArgName("x.y.z"), -2.5, int(now32+14)),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, -2, -3, 4, nan, 0, nan, 5, 6},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					fmt.Sprintf("scaleAfterTimestamp(x.y.z,-2.5,%d)", now32+14),
					[]float64{1, -2, -3, -10, nan, 0, nan, -12.5, -15},
					5,
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

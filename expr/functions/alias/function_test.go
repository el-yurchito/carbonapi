package alias

import (
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

func TestAlias(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("alias", "metric1", parser.ArgValue("renamed")),
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric1",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"renamed",
					[]float64{1, 2, 3, 4, 5},
					1,
					now32,
				),
			},
			nil,
		},
		{
			parser.NewExpr("alias", "metric2", parser.ArgValue("some format ${expr} str ${expr} and another ${expr"), parser.ArgName("true")),
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric2",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"some format metric2 str metric2 and another ${expr",
					[]float64{1, 2, 3, 4, 5},
					1,
					now32,
				),
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

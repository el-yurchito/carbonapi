package event

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

func TestEvent(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("event", "metric1"),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1",
					[]float64{1, math.NaN(), math.NaN(), 3, 4, 12, 15}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("event(metric1)",
				[]float64{1, 0, 0, 3, 4, 12, 15}, 1, now32)},
			nil,
		},

		{
			parser.NewExpr("event", "metric[12]"),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 1, 3, 5, 8, 13, 21}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 1, 1, 1, 1, 1, 1}, 1, now32),
				},
			},
			nil,
			types.ErrWildcardNotAllowed,
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt, true)
		})
	}

}

func TestEventNoData(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tt := th.EvalTestItem{
		parser.NewExpr("event", "metric4"),
		map[parser.MetricRequest][]*types.MetricData{},
		[]*types.MetricData{types.MakeMetricData("event(metric4)",
			[]float64{0, 0}, 30, now32)},
		nil,
	}

	testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
	t.Run(testName, func(t *testing.T) {
		tt.Want[0] = types.MakeMetricData("event(metric4)", []float64{0, 0}, 30, now32)
		th.TestEvalExprWithLimits(t, &tt, true, 0, 59)

		tt.Want[0] = types.MakeMetricData("event(metric4)", []float64{0, 0, 0}, 30, now32)
		th.TestEvalExprWithLimits(t, &tt, true, 0, 60)
	})

}

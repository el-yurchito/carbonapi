package groupLeftTagged

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/tests"
)

var (
	now32 = int32(time.Now().Unix())
	vals1 = []float64{1.1, 1.2, 1.3, 1.4, 1.5}
	vals2 = []float64{2.1, 2.2, 2.3, 2.4, 2.5}
)

func init() {
	md := New("")
	evaluator := tests.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func Test_groupLeftTagged_Do(t *testing.T) {
	testCases := []tests.EvalTestItem{
		{
			// tags aren't specified, join by name
			E: parser.NewExpr("groupLeftTagged", "*.a", "*.b"),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("aaa;tag1=val11", vals1, 1, now32),
					types.MakeMetricData("aaa;tag1=val12", vals1, 1, now32),
					types.MakeMetricData("aaa;tag1=val13", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("aaa;tag2=val21", vals2, 1, now32),
					types.MakeMetricData("aaa;tag2=val22", vals2, 1, now32),
					types.MakeMetricData("bbb;tag2=val23", vals2, 1, now32), // will not be matched
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("aaa;tag1=val11;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("aaa;tag1=val11;tag2=val22", vals1, 1, now32),
				types.MakeMetricData("aaa;tag1=val12;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("aaa;tag1=val12;tag2=val22", vals1, 1, now32),
				types.MakeMetricData("aaa;tag1=val13;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("aaa;tag1=val13;tag2=val22", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			// join by `tagJoin` value
			E: parser.NewExpr("groupLeftTagged", "*.a", "*.b", parser.ArgValue("tagJoin")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a1;tag1=val11;tagJoin=aaa", vals1, 1, now32),
					types.MakeMetricData("a2;tag1=val12;tagJoin=aaa", vals1, 1, now32),
					types.MakeMetricData("a3;tag1=val13;tagJoin=bbb", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("b1;tagJoin=aaa;tag2=val21", vals2, 1, now32),
					types.MakeMetricData("b2;tagJoin=aaa;tag2=val22", vals2, 1, now32),
					types.MakeMetricData("b3;tagJoin=aaa;tag2=val23", vals2, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("a1;tag1=val11;tagJoin=aaa;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("a1;tag1=val11;tagJoin=aaa;tag2=val22", vals1, 1, now32),
				types.MakeMetricData("a1;tag1=val11;tagJoin=aaa;tag2=val23", vals1, 1, now32),
				types.MakeMetricData("a2;tag1=val12;tagJoin=aaa;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("a2;tag1=val12;tagJoin=aaa;tag2=val22", vals1, 1, now32),
				types.MakeMetricData("a2;tag1=val12;tagJoin=aaa;tag2=val23", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			// join by `tagJoin` value
			E: parser.NewExpr("groupLeftTagged", "*.a", "*.b", parser.ArgValue("tagJoin")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a1;tag1=val11;tagJoin=aaa", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("b1;tag1=val21;tagJoin=aaa", vals2, 1, now32),
				},
			},
			Want: []*types.MetricData{
				// joined series has the same tag `tag1`, last value wins
				types.MakeMetricData("a1;tag1=val21;tagJoin=aaa", vals1, 1, now32),
			},
			WantError: nil,
		},
	}

	for _, tc := range testCases {
		testName := tc.E.Target() + "(" + tc.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			tests.TestEvalExpr(t, &tc, true)
		})
	}
}

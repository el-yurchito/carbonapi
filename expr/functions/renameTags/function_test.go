package renameTags

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

func Test_renameTags_Do(t *testing.T) {
	testCases := []tests.EvalTestItem{
		{
			E: parser.NewExpr(
				"renameTags", "*.a",
				parser.ArgValue("tag1"), parser.ArgValue("newTag1"),
				parser.ArgValue("tag4"), parser.ArgValue("newTag4"),
			),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a1;tag1=val11;tag2=val21", vals1, 1, now32),
					types.MakeMetricData("a2;tag1=val21;tag2=val22", vals1, 1, now32),
					types.MakeMetricData("a3;tag4=val34;tag2=val32", vals1, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("a1;newTag1=val11;tag2=val21", vals1, 1, now32),
				types.MakeMetricData("a2;newTag1=val21;tag2=val22", vals1, 1, now32),
				types.MakeMetricData("a3;newTag4=val34;tag2=val32", vals1, 1, now32),
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

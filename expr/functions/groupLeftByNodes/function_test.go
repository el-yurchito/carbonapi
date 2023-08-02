package groupLeftByNodes

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/tests"
)

var (
	nan   = math.NaN()
	now32 = int32(time.Now().Unix())
	vals1 = []float64{1.1, 1.2, 1.3, 1.4, 1.5}
	vals2 = []float64{2.1, math.NaN(), 2.3, math.NaN(), math.NaN()}
	vals3 = []float64{1.1, math.NaN(), 1.3, math.NaN(), math.NaN()}
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

func Test_groupLeftByNodes_Do(t *testing.T) {
	testCases := []tests.EvalTestItem{
		{
			E: parser.NewExpr("groupLeftByNodes", "*.a", "*.b", parser.ArgValue("default"), 1, -1),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a11.xxx.yyy", vals1, 1, now32),
					types.MakeMetricData("a21.xxx.yyy", vals1, 1, now32),
					types.MakeMetricData("a31.xxx.yyy", vals1, 1, now32),
					types.MakeMetricData("a41.xxx.yyy", vals1, 1, now32),
					types.MakeMetricData("a51.zzz.ttt", vals1, 1, now32),
					types.MakeMetricData("a61.yyy.xxx", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("a12.xxx.yyy", vals2, 1, now32),
					types.MakeMetricData("a22.zzz.ttt", vals2, 1, now32),
					types.MakeMetricData("a32.kkk.mmm", vals2, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("a11.xxx.yyy.a12.xxx.yyy", vals1, 1, now32),
				types.MakeMetricData("a21.xxx.yyy.a12.xxx.yyy", vals1, 1, now32),
				types.MakeMetricData("a31.xxx.yyy.a12.xxx.yyy", vals1, 1, now32),
				types.MakeMetricData("a41.xxx.yyy.a12.xxx.yyy", vals1, 1, now32),
				types.MakeMetricData("a51.zzz.ttt.a22.zzz.ttt", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			// empty nodes list - incorrect
			E: parser.NewExpr("groupLeftByNodes", "*.a", "*.b", parser.ArgValue("default")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {},
				{"*.b", 0, 1}: {},
			},
			Want:      nil,
			WantError: fmt.Errorf("at least one node must be specified"),
		},
		{
			// unknown parameter "valuesFrom"
			E: parser.NewExpr("groupLeftByNodes", "*.a", "*.b", parser.ArgValue("fgsfds")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {},
				{"*.b", 0, 1}: {},
			},
			Want:      nil,
			WantError: fmt.Errorf("unknown \"fgsfds\" value: only \"default\" and \"common\" are allowed"),
		},
		{
			E: parser.NewExpr("groupLeftByNodes", "*.a", "*.b", parser.ArgValue("default"), 1),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a11.join.c11", vals1, 1, now32),
					types.MakeMetricData("a21.join.c21", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("a12.join.c12", vals2, 1, now32),
					types.MakeMetricData("a22.join.c22", vals2, 1, now32),
					types.MakeMetricData("a32.join.c32", vals2, 1, now32),
					types.MakeMetricData("a42.join.c42", vals2, 1, now32),
					types.MakeMetricData("a52.not-join.c52", vals2, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("a11.join.c11.a12.join.c12", vals1, 1, now32),
				types.MakeMetricData("a11.join.c11.a22.join.c22", vals1, 1, now32),
				types.MakeMetricData("a11.join.c11.a32.join.c32", vals1, 1, now32),
				types.MakeMetricData("a11.join.c11.a42.join.c42", vals1, 1, now32),
				types.MakeMetricData("a21.join.c21.a12.join.c12", vals1, 1, now32),
				types.MakeMetricData("a21.join.c21.a22.join.c22", vals1, 1, now32),
				types.MakeMetricData("a21.join.c21.a32.join.c32", vals1, 1, now32),
				types.MakeMetricData("a21.join.c21.a42.join.c42", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			E: parser.NewExpr("groupLeftByNodes", "*.a", "*.b", parser.ArgValue("common"), 1),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("xxx.aaa", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("yyy.aaa", vals2, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("xxx.aaa.yyy.aaa", vals3, 1, now32),
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

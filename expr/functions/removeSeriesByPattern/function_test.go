package removeSeriesByPattern

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/tests"
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

func Test_removeSeriesByPattern(t *testing.T) {
	now32 := int32(time.Now().Unix())
	vals1 := []float64{1.1, 1.2, 1.3, 1.4, 1.5}
	vals2 := []float64{2.1, 2.2, 2.3, 2.4, 2.5}
	vals3 := []float64{3.1, 3.2, 3.3, 3.4, 3.5}
	vals4 := []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}

	testCases := []tests.EvalTestItem{
		{
			E: parser.NewExpr("removeSeriesByPattern", "a*.*.ccc", "*"),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"a*.*.ccc", 0, 1}: {
					types.MakeMetricData("aa.b1.ccc", vals1, 1, now32),
					types.MakeMetricData("aaa.b2.ccc", vals2, 1, now32),
					types.MakeMetricData("aaaa.b3.ccc", vals3, 1, now32),
				},
				{"*", 0, 1}: {
					types.MakeMetricData("aa", vals4, 1, now32),  // matches `aa.b1.ccc` series
					types.MakeMetricData("b2", vals4, 1, now32),  // matches `aaa.b2.ccc` series
					types.MakeMetricData("cde", vals4, 1, now32), // matches no series
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("aaaa.b3.ccc", vals3, 1, now32),
			},
			WantError: nil,
		},
		{
			E: parser.NewExpr("removeSeriesByPattern", "a*.*.ccc", "*"),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"a*.*.ccc", 0, 1}: {
					types.MakeMetricData("aa.b1.ccc", vals1, 1, now32),
					types.MakeMetricData("aaa.b2.ccc", vals2, 1, now32),
					types.MakeMetricData("aaaa.b3.ccc", vals3, 1, now32),
				},
				{"*", 0, 1}: {
					types.MakeMetricData("xyz", vals4, 1, now32), // matches no series
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("aa.b1.ccc", vals1, 1, now32),
				types.MakeMetricData("aaa.b2.ccc", vals2, 1, now32),
				types.MakeMetricData("aaaa.b3.ccc", vals3, 1, now32),
			},
			WantError: nil,
		},
		{
			E: parser.NewExpr("removeSeriesByPattern", "a*.*.ccc", "*"),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"a*.*.ccc", 0, 1}: {
					types.MakeMetricData("aa.b1.ccc", vals1, 1, now32),
					types.MakeMetricData("aaa.b2.ccc", vals2, 1, now32),
					types.MakeMetricData("aaaa.b3.ccc", vals3, 1, now32),
				},
				{"*", 0, 1}: {
					types.MakeMetricData("ccc", vals4, 1, now32), // matches every series
				},
			},
			Want:      []*types.MetricData{},
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

package groupByTags

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/functions/sum"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	s := sum.New("")
	for _, m := range s {
		metadata.RegisterFunction(m.Name, m.F)
	}
	md := New("")
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}

	evaluator := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
}

func TestGroupByTags(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			parser.NewExpr("groupByTags",
				"metric1.foo.*",
				parser.ArgValue("sum"),
				parser.ArgValue("dc"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*", 0, 1}: {
					types.MakeMetricData("metric1.foo;cpu=cpu1;dc=dc1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu2;dc=dc1", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu3;dc=dc1", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu4;dc=dc1", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"groupByTags",
			map[string][]*types.MetricData{
				"dc1": {types.MakeMetricData("dc1", []float64{25, 29, 33, 37, 41}, 1, now32)},
			},
		},
		{
			parser.NewExpr("groupByTags",
				"metric1.foo.*",
				parser.ArgValue("sum"),
				parser.ArgValue("dc"),
				parser.ArgValue("cpu"),
				parser.ArgValue("rack"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*", 0, 1}: {
					types.MakeMetricData("metric1.foo;cpu=cpu1;dc=dc1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu2;dc=dc1", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu3;dc=dc1", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu4;dc=dc1", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"groupByTags",
			map[string][]*types.MetricData{
				"dc1.cpu1.": {types.MakeMetricData("dc1.cpu1.", []float64{1, 2, 3, 4, 5}, 1, now32)},
				"dc1.cpu2.": {types.MakeMetricData("dc1.cpu2.", []float64{6, 7, 8, 9, 10}, 1, now32)},
				"dc1.cpu3.": {types.MakeMetricData("dc1.cpu3.", []float64{11, 12, 13, 14, 15}, 1, now32)},
				"dc1.cpu4.": {types.MakeMetricData("dc1.cpu4.", []float64{7, 8, 9, 10, 11}, 1, now32)},
			},
		},
		{
			parser.NewExpr("groupByTags",
				"summarize(x.y.*, 1min, sum, false)",
				parser.ArgValue("sum"),
				parser.ArgValue("key2"),
				parser.ArgValue("key1"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"summarize(x.y.*, 1min, sum, false)", 0, 1}: {
					types.MakeMetricData("summarize(x.y.z1;key1=value11;key2=value21, 1min, sum, false)", []float64{1, 1, 1}, 1, now32),
					types.MakeMetricData("summarize(x.y.z2;key1=value12;key2=value22, 1min, sum, false)", []float64{2, 2, 2}, 1, now32),
					types.MakeMetricData("summarize(x.y.z3;key1=value13;key2=value23, 1min, sum, false)", []float64{3, 3, 3}, 1, now32),
				},
			},
			"groupByTags",
			map[string][]*types.MetricData{
				"value21.value11": {types.MakeMetricData("value21.value11", []float64{1, 1, 1}, 1, now32)},
				"value22.value12": {types.MakeMetricData("value22.value12", []float64{2, 2, 2}, 1, now32)},
				"value23.value13": {types.MakeMetricData("value23.value13", []float64{3, 3, 3}, 1, now32)},
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

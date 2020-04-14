package sortByName

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

func TestSortByName(t *testing.T) {
	now32 := int32(time.Now().Unix())

	testCases := []th.EvalTestItem{
		{
			parser.NewExpr("sortByName", parser.ArgName("metric.foo.*")),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // 100 is placed between 1 and 2 because it is alphabetical sort
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
			},
			nil,
		},
		{
			parser.NewExpr("sortByName", parser.ArgName("metric.foo.*"), parser.NewValueExpr("true")),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // "natural" sort method considers that metrics contain numbers
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
			},
			nil,
		},
		{
			parser.NewExpr("sortByName", parser.ArgName("metric.foo.*"), parser.NewValueExpr("false"), parser.NewValueExpr("true")),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // alphabetical reverse sort
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
			},
			nil,
		},
		{
			parser.NewExpr("sortByName", parser.ArgName("metric.foo.*"), parser.NewValueExpr("true"), parser.NewValueExpr("true")),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // "natural" reverse sort
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
			},
			nil,
		},
	}

	for _, testCase := range testCases {
		testName := testCase.E.Target() + "(" + testCase.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase, true)
		})
	}
}

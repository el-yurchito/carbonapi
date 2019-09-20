package timeShiftByMetric

import (
	"fmt"
	"math"
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

func TestTimeShift(t *testing.T) {
	nan := math.NaN()
	now32 := int32(time.Now().Unix())

	testCases := []th.EvalTestItem{
		// 1. Versions: 1_0, 1_1, 1_2, 1_3, 2_0, 2_1, 2_2, 3_0, 3_1. Each two consequential versions are have 1 time unit between them.
		// 2. Leading versions: 1_3, 2_2, 3_1.
		// 3. Version 2_2 is 2 time units behind version 3_1. Version 1_3 is 3 time units behind version 2_2 therefore 5 units behind version 3_1.
		th.EvalTestItem{
			parser.NewExpr("timeShiftByMetric",
				"apps.*.metric", "apps.mark.*", 1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"apps.*.metric", 0, 1}: []*types.MetricData{
					types.MakeMetricData("apps.1_3.metric", []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, nan, nan}, 1, now32),
					types.MakeMetricData("apps.2.metric", []float64{nan, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, nan}, 1, now32),
					types.MakeMetricData("apps.3.metric", []float64{nan, nan, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9}, 1, now32),
				},
				parser.MetricRequest{"apps.mark.*", 0, 1}: []*types.MetricData{
					// leading versions
					types.MakeMetricData("apps.mark.1_3", []float64{nan, nan, nan, 1, nan, nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.2_2", []float64{nan, nan, nan, nan, nan, nan, 1, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.3_1", []float64{nan, nan, nan, nan, nan, nan, nan, nan, 1, nan, nan}, 1, now32),
					// rest
					types.MakeMetricData("apps.mark.1_0", []float64{1, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.1_1", []float64{nan, 1, nan, nan, nan, nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.1_2", []float64{nan, nan, 1, nan, nan, nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.2_0", []float64{nan, nan, nan, nan, 1, nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.2_1", []float64{nan, nan, nan, nan, nan, 1, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.3_0", []float64{nan, nan, nan, nan, nan, nan, nan, 1, nan, nan, nan}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("timeShiftByMetric(apps.1_3.metric)", []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, nan, nan}, 1, now32+5),
				types.MakeMetricData("timeShiftByMetric(apps.2.metric)", []float64{nan, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, nan}, 1, now32+2),
				types.MakeMetricData("timeShiftByMetric(apps.3.metric)", []float64{nan, nan, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9}, 1, now32),
			},
			nil,
		},

		th.EvalTestItem{
			// 1. Versions: 1_0, 1_1, 2_0.
			// 2. Leading versions: 1_1, 2_0.
			// 3. Version 1_1 is 4 time units behind versions 2_0.
			parser.NewExpr("timeShiftByMetric",
				"*.metric", "apps.mark.*", 0,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"*.metric", 0, 1}: []*types.MetricData{
					types.MakeMetricData("1_1.metric", []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7}, 1, now32),
					types.MakeMetricData("2_0.metric", []float64{2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7}, 1, now32),
				},
				parser.MetricRequest{"apps.mark.*", 0, 1}: []*types.MetricData{
					// leading versions
					types.MakeMetricData("apps.mark.1_1", []float64{nan, nan, 1, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.2_0", []float64{nan, nan, nan, nan, nan, nan, 1}, 1, now32),
					// rest
					types.MakeMetricData("apps.mark.1_0", []float64{1, nan, nan, nan, nan, nan, nan}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("timeShiftByMetric(1_1.metric)", []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7}, 1, now32+4),
				types.MakeMetricData("timeShiftByMetric(2_0.metric)", []float64{2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7}, 1, now32),
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

func TestBadMarks(t *testing.T) {
	nan := math.NaN()
	now32 := int32(time.Now().Unix())

	testCases := []th.EvalTestItem{
		// we have only one major version here, it's not right
		th.EvalTestItem{
			parser.NewExpr("timeShiftByMetric",
				"apps.*.metric", "apps.mark.*", 1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"apps.*.metric", 0, 1}: []*types.MetricData{
					types.MakeMetricData("apps.1.metric", []float64{1, 2, 3}, 1, now32),
					types.MakeMetricData("apps.2.metric", []float64{1, 2, 3}, 1, now32),
					types.MakeMetricData("apps.3.metric", []float64{1, 2, 3}, 1, now32),
				},
				parser.MetricRequest{"apps.mark.*", 0, 1}: []*types.MetricData{
					types.MakeMetricData("apps.mark.1_0", []float64{1, nan, nan}, 1, now32),
					types.MakeMetricData("apps.mark.1_1", []float64{nan, 1, nan}, 1, now32),
					types.MakeMetricData("apps.mark.1_2", []float64{nan, nan, 1}, 1, now32),
				},
			},
			nil,
			fmt.Errorf("bad data: could not find 2 marks, only 1 found"),
		},
	}

	for _, testCase := range testCases {
		testName := testCase.E.Target() + "(" + testCase.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase, true)
		})
	}
}

func TestNotEnoughSeries(t *testing.T) {
	nan := math.NaN()
	now32 := int32(time.Now().Unix())
	testCases := make([]th.EvalTestItem, 0, 4)

	// enough metrics but not enough marks
	for i := 0; i < 2; i++ {
		marksData := make([]*types.MetricData, 0, 1)
		for j := 0; j < i; j++ {
			marksData = append(marksData, types.MakeMetricData("apps.mark.1_0", []float64{1, nan, nan, nan, nan}, 1, now32))
		}

		metricsData := []*types.MetricData{
			types.MakeMetricData("apps.1.metric", []float64{1, 2, 3, 4, 5}, 1, now32),
			types.MakeMetricData("apps.2.metric", []float64{1, 2, 3, 4, 5}, 1, now32),
		}

		testCases = append(testCases, th.EvalTestItem{
			parser.NewExpr("timeShiftByMetric",
				"apps.*.metric", "apps.mark.*", 1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"apps.*.metric", 0, 1}: metricsData,
				parser.MetricRequest{"apps.mark.*", 0, 1}:   marksData,
			},
			nil,
			fmt.Errorf("bad data: need at least 2 %s data sets to process, got %d", "marks", i),
		})
	}

	// enough marks but not enough metrics
	for i := 0; i < 2; i++ {
		metricsData := make([]*types.MetricData, 0, 1)
		for j := 0; j < i; j++ {
			metricsData = append(metricsData, types.MakeMetricData("apps.1.metric", []float64{1, 2, 3, 4, 5}, 1, now32))
		}

		marksData := []*types.MetricData{
			types.MakeMetricData("apps.mark.1_0", []float64{1, nan}, 1, now32),
			types.MakeMetricData("apps.mark.2_0", []float64{nan, 2}, 1, now32),
		}

		testCases = append(testCases, th.EvalTestItem{
			parser.NewExpr("timeShiftByMetric",
				"apps.*.metric", "apps.mark.*", 1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"apps.*.metric", 0, 1}: metricsData,
				parser.MetricRequest{"apps.mark.*", 0, 1}:   marksData,
			},
			nil,
			fmt.Errorf("bad data: need at least 2 %s data sets to process, got %d", "metrics", i),
		})
	}

	for _, testCase := range testCases {
		testName := testCase.E.Target() + "(" + testCase.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase, true)
		})
	}
}

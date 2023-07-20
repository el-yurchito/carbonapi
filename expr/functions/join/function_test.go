package join

import (
	"fmt"
	"reflect"
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

func Test_join_simpleValidCalls(t *testing.T) {
	testCases := []tests.EvalTestItem{
		{
			E: parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("AND")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a1", vals1, 1, now32),
					types.MakeMetricData("a2", vals1, 1, now32),
					types.MakeMetricData("a3", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("a1", vals2, 1, now32),
					types.MakeMetricData("a2", vals2, 1, now32),
					types.MakeMetricData("a4", vals2, 1, now32),
				},
			},
			// metric names aren't transformed
			// [a1,a2,a3] AND [a1,a2,a4] == [a1,a2]
			Want: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			E: parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("AND"), parser.ArgValue("1"), parser.ArgValue("0")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("ignored.a1", vals1, 1, now32),
					types.MakeMetricData("ignored.a2", vals1, 1, now32),
					types.MakeMetricData("ignored.a3", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("a1", vals2, 1, now32),
					types.MakeMetricData("a2", vals2, 1, now32),
					types.MakeMetricData("a4", vals2, 1, now32),
				},
			},
			// nodes with index 1 are taken from seriesA: [a1,a2,a3]
			// nodes with index 0 are takes from seriesB: [a1,a2,a4]
			// [a1,a2,a3] AND [a1,a2,a4] == [a1,a2]
			// result metric names are presented without transformation
			Want: []*types.MetricData{
				types.MakeMetricData("ignored.a1", vals1, 1, now32),
				types.MakeMetricData("ignored.a2", vals1, 1, now32),
			},
			WantError: nil,
		},
		{
			E: parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("OR"), parser.ArgValue("0"), parser.ArgValue("-1")),
			M: map[parser.MetricRequest][]*types.MetricData{
				{"*.a", 0, 1}: {
					types.MakeMetricData("a1.ignored1", vals1, 1, now32),
					types.MakeMetricData("a2.ignored2", vals1, 1, now32),
				},
				{"*.b", 0, 1}: {
					types.MakeMetricData("also-ignored1.a1", vals2, 1, now32),
					types.MakeMetricData("also-ignored2.a2", vals2, 1, now32),
					types.MakeMetricData("also-ignored3.a3", vals2, 1, now32),
				},
			},
			// nodes with index 0 are taken from seriesA: [a1,a2]
			// nodes with index -1 are takes from seriesB: [a1,a2,a3]
			// [a1,a2] OR [a1,a2,a3] == [a1,a2,a3]
			// result metric names are presented without transformation
			// metric values missing at seriesA are taken from seriesB
			Want: []*types.MetricData{
				types.MakeMetricData("a1.ignored1", vals1, 1, now32),
				types.MakeMetricData("a2.ignored2", vals1, 1, now32),
				types.MakeMetricData("also-ignored3.a3", vals2, 1, now32),
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

func Test_join_invalidArgs(t *testing.T) {
	testCases := []tests.EvalTestItem{
		{
			E:         parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("WRONG_VALUE")),
			M:         nil,
			Want:      nil,
			WantError: fmt.Errorf("unknown join type: WRONG_VALUE"),
		},
		{
			E:         parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("AND"), parser.ArgValue("NOT_INT")),
			M:         nil,
			Want:      nil,
			WantError: fmt.Errorf("failed to parse node numbers list 'NOT_INT': strconv.Atoi: parsing \"NOT_INT\": invalid syntax"),
		},
		{
			E:         parser.NewExpr("join", "*.a", "*.b", parser.ArgValue("AND"), parser.ArgValue("0 ")),
			M:         nil,
			Want:      nil,
			WantError: fmt.Errorf("failed to parse node numbers list '0 ': strconv.Atoi: parsing \"0 \": invalid syntax"),
		},
	}

	for _, tc := range testCases {
		testName := tc.E.Target() + "(" + tc.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			tests.TestEvalExpr(t, &tc, true)
		})
	}
}

func Test_nodeNumbers_transform(t *testing.T) {
	type testCase struct {
		nodeNumbersSource string
		namesOriginal     []string
		resultExpected    []string
	}

	testCases := []testCase{
		{
			nodeNumbersSource: "0",
			namesOriginal:     []string{"a1", "a2.b2", "a3.b3.c3"},
			resultExpected:    []string{"a1", "a2", "a3"},
		},
		{
			nodeNumbersSource: "1",
			namesOriginal:     []string{"a1", "a2.b2", "a3.b3.c3"},
			resultExpected:    []string{"", "b2", "b3"},
		},
		{
			nodeNumbersSource: "-1.0.1",
			namesOriginal:     []string{"a1.b1.c1", "a2.b2.c2", "a3.b3.c3"},
			resultExpected:    []string{"c1.a1.b1", "c2.a2.b2", "c3.a3.b3"},
		},
		{
			nodeNumbersSource: "-3.2",
			namesOriginal:     []string{"a1.b1.c1.d1", "a2.b2.c2.d2", "a3.b3.c3.d3"},
			resultExpected:    []string{"b1.c1", "b2.c2", "b3.c3"},
		},
		{
			nodeNumbersSource: "-3.1",
			namesOriginal:     []string{"a1.b1.c1.d1", "a2.b2"},
			resultExpected:    []string{"b1.b1", ".b2"}, // missing index: negative
		},
		{
			nodeNumbersSource: "2",
			namesOriginal:     []string{"a1.b1.c1.d1", "a2.b2.c2", "a3.b3"},
			resultExpected:    []string{"c1", "c2", ""}, // missing index: positive
		},
		{
			nodeNumbersSource: "2.3",
			namesOriginal:     []string{"a1.b1.c1.d1", "a2.b2.c2.d2", "a3.b3.c3"},
			resultExpected:    []string{"c1.d1", "c2.d2", "c3."}, // missing index: positive
		},
	}

	for _, tc := range testCases {
		tcName := fmt.Sprintf(
			"testCase{source:'%s',names:%v,expected:%v}",
			tc.nodeNumbersSource, tc.namesOriginal, tc.resultExpected,
		)
		t.Run(tcName, func(t *testing.T) {
			if len(tc.namesOriginal) != len(tc.resultExpected) {
				t.Fatal("namesOriginal and resultExpected must have the same length")
			}

			transformer, err := parseNodesList(tc.nodeNumbersSource)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			for i, name := range tc.namesOriginal {
				resultActual := transformer.transform(name)
				if resultActual != tc.resultExpected[i] {
					t.Fatalf(
						"mismatch at index %d: want '%s', have '%s'",
						i, tc.resultExpected[i], resultActual,
					)
				}
			}
		})
	}
}

func Test_operations(t *testing.T) {
	type operation func(seriesA, seriesB []*types.MetricData, transformerA, transformerB metricNameTransformer) []*types.MetricData

	type testCase struct {
		seriesA       []*types.MetricData
		seriesB       []*types.MetricData
		expected      []*types.MetricData
		operationFn   operation
		operationName string
	}

	testCases := []testCase{
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("a1", vals2, 1, now32),
				types.MakeMetricData("a2", vals2, 1, now32),
			},
			expected: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
			},
			operationFn:   doAnd,
			operationName: "doAnd",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("b1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			expected:      []*types.MetricData{},
			operationFn:   doAnd,
			operationName: "doAnd",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("b1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			expected: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
				types.MakeMetricData("b1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			operationFn:   doOr,
			operationName: "doOr",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("a1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			// a1 is presented in both seriesA and seriesB
			// its values will be taken from the first operand (seriesA)
			expected: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			operationFn:   doOr,
			operationName: "doOr",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("b1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			expected: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
				types.MakeMetricData("b1", vals2, 1, now32),
				types.MakeMetricData("b2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			operationFn:   doXor,
			operationName: "doXor",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("b3", vals1, 1, now32),
				types.MakeMetricData("a4", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("a1", vals2, 1, now32),
				types.MakeMetricData("a2", vals2, 1, now32),
				types.MakeMetricData("b3", vals2, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			// a1, a2 and b3 are presented in both seriesA and seriesB
			// so only a4 and b4 are going to form the result
			expected: []*types.MetricData{
				types.MakeMetricData("a4", vals1, 1, now32),
				types.MakeMetricData("b4", vals2, 1, now32),
			},
			operationFn:   doXor,
			operationName: "doXor",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("a1", vals2, 1, now32),
				types.MakeMetricData("a2", vals2, 1, now32),
			},
			expected: []*types.MetricData{
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			operationFn:   doSub,
			operationName: "doSub",
		},
		{
			seriesA: []*types.MetricData{
				types.MakeMetricData("a1", vals1, 1, now32),
				types.MakeMetricData("a2", vals1, 1, now32),
				types.MakeMetricData("a3", vals1, 1, now32),
			},
			seriesB: []*types.MetricData{
				types.MakeMetricData("a1", vals2, 1, now32),
				types.MakeMetricData("a2", vals2, 1, now32),
				types.MakeMetricData("a3", vals2, 1, now32),
				types.MakeMetricData("a4", vals2, 1, now32),
			},
			expected:      []*types.MetricData{},
			operationFn:   doSub,
			operationName: "doSub",
		},
	}

	for _, tc := range testCases {
		tcName := fmt.Sprintf(
			"testCase{operation=%s,a=%v,b=%v,expected=%v}",
			tc.operationName, pickMetricNames(tc.seriesA), pickMetricNames(tc.seriesB), pickMetricNames(tc.expected),
		)
		t.Run(tcName, func(t *testing.T) {
			actual := tc.operationFn(tc.seriesA, tc.seriesB, noop{}, noop{})
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("wrong result: want %v, have %v", tc.expected, actual)
			}
		})
	}
}

func pickMetricNames(md []*types.MetricData) []string {
	result := make([]string, 0, len(md))
	for i := range md {
		result = append(result, md[i].Name)
	}
	return result
}

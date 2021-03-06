package tests

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type FuncEvaluator struct {
	eval func(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

func (evaluator *FuncEvaluator) EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.IsName() {
		return values[parser.MetricRequest{Metric: e.Target(), From: from, Until: until}], nil
	} else if e.IsConst() {
		p := types.MetricData{FetchResponse: types.FetchResponse{Name: e.Target(), Values: []float64{e.FloatValue()}}}
		return []*types.MetricData{&p}, nil
	}
	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.Args()) == 0 {
		return nil, parser.ErrMissingArgument
	}

	return evaluator.eval(e, from, until, values)
}

func EvaluatorFromFunc(function interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: function.Do,
	}

	return e
}

func EvaluatorFromFuncWithMetadata(metadata map[string]interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: func(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
			if f, ok := metadata[e.Target()]; ok {
				return f.Do(e, from, until, values)
			}
			return nil, fmt.Errorf("unknown function: %v", e.Target())
		},
	}
	return e
}

func DeepClone(original map[parser.MetricRequest][]*types.MetricData) map[parser.MetricRequest][]*types.MetricData {
	clone := map[parser.MetricRequest][]*types.MetricData{}
	for key, originalMetrics := range original {
		copiedMetrics := make([]*types.MetricData, 0, len(originalMetrics))
		for _, originalMetric := range originalMetrics {
			copiedMetric := types.MetricData{
				FetchResponse: types.FetchResponse{
					Name:      originalMetric.Name,
					StartTime: originalMetric.StartTime,
					StopTime:  originalMetric.StopTime,
					StepTime:  originalMetric.StepTime,
					Values:    make([]float64, len(originalMetric.Values)),
					IsAbsent:  make([]bool, len(originalMetric.IsAbsent)),
				},
			}

			copy(copiedMetric.Values, originalMetric.Values)
			copy(copiedMetric.IsAbsent, originalMetric.IsAbsent)
			copiedMetrics = append(copiedMetrics, &copiedMetric)
		}

		clone[key] = copiedMetrics
	}

	return clone
}

func DeepEqual(t *testing.T, target string, original, modified map[parser.MetricRequest][]*types.MetricData) {
	for key := range original {
		if len(original[key]) == len(modified[key]) {
			for i := range original[key] {
				if !reflect.DeepEqual(original[key][i], modified[key][i]) {
					t.Errorf(
						"%s: source data was modified key %v index %v original:\n%v\n modified:\n%v",
						target,
						key,
						i,
						original[key][i],
						modified[key][i],
					)
				}
			}
		} else {
			t.Errorf(
				"%s: source data was modified key %v original length %d, new length %d",
				target,
				key,
				len(original[key]),
				len(modified[key]),
			)
		}
	}
}

const eps = 0.0000000001

func NearlyEqual(a []float64, absent []bool, b []float64) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		// "same"
		if absent[i] && math.IsNaN(b[i]) {
			continue
		}
		if absent[i] || math.IsNaN(b[i]) {
			// unexpected NaN
			return false
		}
		// "close enough"
		if math.Abs(v-b[i]) > eps {
			return false
		}
	}

	return true
}

func NearlyEqualMetrics(a, b *types.MetricData) bool {

	if len(a.IsAbsent) != len(b.IsAbsent) {
		return false
	}

	for i := range a.IsAbsent {
		if a.IsAbsent[i] != b.IsAbsent[i] {
			return false
		}
		// "close enough"
		if math.Abs(a.Values[i]-b.Values[i]) > eps {
			return false
		}
	}

	return true
}

type SummarizeEvalTestItem struct {
	E     parser.Expr
	M     map[parser.MetricRequest][]*types.MetricData
	W     []float64
	Name  string
	Step  int32
	Start int32
	Stop  int32
}

func InitTestSummarize() (int32, int32, int32) {
	t0, err := time.Parse(time.UnixDate, "Wed Sep 10 10:32:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirtyTwo := int32(t0.Unix())

	t0, err = time.Parse(time.UnixDate, "Wed Sep 10 10:59:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenFiftyNine := int32(t0.Unix())

	t0, err = time.Parse(time.UnixDate, "Wed Sep 10 10:30:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirty := int32(t0.Unix())

	return tenThirtyTwo, tenFiftyNine, tenThirty
}

func TestSummarizeEvalExpr(t *testing.T, tt *SummarizeEvalTestItem) {
	evaluator := metadata.GetEvaluator()

	t.Run(tt.Name, func(t *testing.T) {
		originalMetrics := DeepClone(tt.M)
		g, err := evaluator.EvalExpr(tt.E, 0, 1, tt.M)
		if err != nil {
			t.Errorf("failed to eval %v: %+v", tt.Name, err)
			return
		}
		DeepEqual(t, g[0].Name, originalMetrics, tt.M)
		if g[0].StepTime != tt.Step {
			t.Errorf("bad Step for %s:\ngot  %d\nwant %d", g[0].Name, g[0].StepTime, tt.Step)
		}
		if g[0].StartTime != tt.Start {
			t.Errorf("bad Start for %s: got %s want %s", g[0].Name, time.Unix(int64(g[0].StartTime), 0).Format(time.StampNano), time.Unix(int64(tt.Start), 0).Format(time.StampNano))
		}
		if g[0].StopTime != tt.Stop {
			t.Errorf("bad Stop for %s: got %s want %s", g[0].Name, time.Unix(int64(g[0].StopTime), 0).Format(time.StampNano), time.Unix(int64(tt.Stop), 0).Format(time.StampNano))
		}

		if !NearlyEqual(g[0].Values, g[0].IsAbsent, tt.W) {
			t.Errorf("failed: %s:\ngot  %+v,\nwant %+v", g[0].Name, g[0].Values, tt.W)
		}
		if g[0].Name != tt.Name {
			t.Errorf("bad Name for %+v: got %v, want %v", g, g[0].Name, tt.Name)
		}
	})
}

type MultiReturnEvalTestItem struct {
	E       parser.Expr
	M       map[parser.MetricRequest][]*types.MetricData
	Name    string
	Results map[string][]*types.MetricData
}

func TestMultiReturnEvalExpr(t *testing.T, tt *MultiReturnEvalTestItem) {
	evaluator := metadata.GetEvaluator()

	originalMetrics := DeepClone(tt.M)
	g, err := evaluator.EvalExpr(tt.E, 0, 1, tt.M)
	if err != nil {
		t.Errorf("failed to eval %v: %+v", tt.Name, err)
		return
	}
	DeepEqual(t, tt.Name, originalMetrics, tt.M)
	if len(g) == 0 {
		t.Errorf("returned no data %v", tt.Name)
		return
	}
	if g[0] == nil {
		t.Errorf("returned no value %v", tt.Name)
		return
	}
	if g[0].StepTime == 0 {
		t.Errorf("missing Step for %+v", g)
	}
	if len(g) != len(tt.Results) {
		t.Errorf("unexpected results len: got %d, want %d", len(g), len(tt.Results))
	}
	for _, gg := range g {
		r, ok := tt.Results[gg.Name]
		if !ok {
			t.Errorf("missing result Name: %v", gg.Name)
			continue
		}
		if r[0].Name != gg.Name {
			t.Errorf("result Name mismatch, got\n%#v,\nwant\n%#v", gg.Name, r[0].Name)
		}
		if !reflect.DeepEqual(r[0].Values, gg.Values) || !reflect.DeepEqual(r[0].IsAbsent, gg.IsAbsent) ||
			r[0].StartTime != gg.StartTime ||
			r[0].StopTime != gg.StopTime ||
			r[0].StepTime != gg.StepTime {
			t.Errorf("result mismatch, got\n%#v,\nwant\n%#v", gg, r)
		}
	}
}

type EvalTestItem struct {
	E         parser.Expr
	M         map[parser.MetricRequest][]*types.MetricData
	Want      []*types.MetricData
	WantError error
}

func TestEvalExpr(t *testing.T, tt *EvalTestItem, strictOrder bool) {
	TestEvalExprWithLimits(t, tt, strictOrder, 0, 1)
}

func TestEvalExprWithLimits(t *testing.T, tt *EvalTestItem, strictOrder bool, from, until int32) {
	if (tt.Want == nil) == (tt.WantError == nil) {
		t.Fatalf("Improperly configured: can neither set both nor unset both Want and WantError")
	}

	evaluator := metadata.GetEvaluator()
	originalMetrics := DeepClone(tt.M)
	testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
	actual, err := evaluator.EvalExpr(tt.E, from, until, tt.M)

	if err == nil {
		if tt.WantError != nil {
			t.Fatalf("%s expected error %s but didn't cause any", testName, tt.WantError.Error())
		}
		if len(actual) != len(tt.Want) {
			t.Fatalf("%s returned a different number of metrics, actual %v, Want %v", testName, len(actual), len(tt.Want))
		}
		DeepEqual(t, testName, originalMetrics, tt.M)
		compareMetricDataSets(t, tt, actual, tt.Want, strictOrder)
	} else if tt.WantError == nil {
		t.Fatalf("failed to eval %s: %+v", testName, err)
	} else if err.Error() != tt.WantError.Error() {
		t.Fatalf("%s caused unexpected error: expected:\n%s\nbut was actually:\n%s", testName, tt.WantError.Error(), err.Error())
	}
}

// compareMetricData compares single actual and wanted MetricData objects and returns comparision result
// errorCode == 0 is considered as ok, no errorMessages is provided in this case
// errorCode == 1 is considered as non-fatal failure; errorMessages will be logged, the test case will be considered failed but will continue execution
// errorCode == 2 is considered as fatal failure; errorMessages will be logged, the test case will be considered failed and will stop immediately
func compareMetricData(actual, wanted *types.MetricData) (int, []string) {
	errorCode := 0
	errorMessages := make([]string, 0, 3)

	if actual.StepTime == 0 {
		errorCode = 1
		errorMessages = append(errorMessages, fmt.Sprintf("Missing StepTime for %v", actual))
	}

	if actual.Name != wanted.Name {
		errorCode = 1
		errorMessages = append(errorMessages, fmt.Sprintf("Bad Name for metric: got %s, Want %s", actual.Name, wanted.Name))
	}

	if !NearlyEqualMetrics(actual, wanted) {
		errorCode = 2
		errorMessages = append(errorMessages, fmt.Sprintf("different values for metric %s: got %v, Want %v", actual.Name, actual.Values, wanted.Values))
	}

	return errorCode, errorMessages
}

// compareMetricDataSets compares actual and wanted metric data in the way specified by strictOrder
// if strictOrder is true then data will be compared following the order given
// if strictOrder is false then data will be compared as "set by set",
// i.e. ["aaa", "bbb"] and ["bbb", "aaa"] will be considered equal
func compareMetricDataSets(t *testing.T, tt *EvalTestItem, actual, wanted []*types.MetricData, strictOrder bool) {
	if len(actual) != len(wanted) {
		panic(fmt.Errorf("invalid compareMetricData call: lengths of actual and wanted must be equal"))
	}

	for _, v := range actual {
		if v == nil {
			t.Errorf("returned no value %v", tt.E.RawArgs())
			return
		}
	}
	for i, v := range wanted {
		if v == nil {
			t.Errorf("not specified Want row #%d", i)
			return
		}
	}

	// metric data with their initial position
	type MetricDataOrdered struct {
		data     *types.MetricData
		position int
	}

	// metrics grouped by name
	// wwe assume here that each metric is presented only once
	type MetricGroup map[string]MetricDataOrdered

	mapActual := make(MetricGroup)
	mapWanted := make(MetricGroup)

	for i := 0; i < len(wanted); i++ {
		mapActual[actual[i].Name] = MetricDataOrdered{
			data:     actual[i],
			position: i,
		}
		mapWanted[wanted[i].Name] = MetricDataOrdered{
			data:     wanted[i],
			position: i,
		}
	}

	for keyWanted, metricWanted := range mapWanted {
		if metricActual, ok := mapActual[keyWanted]; !ok {
			t.Errorf("Metric %s is wanted but not presented", keyWanted)
		} else if strictOrder && metricActual.position != metricWanted.position {
			t.Errorf("Wanted metric %s is presented but has wrong position (position is %d, wanted %d)",
				keyWanted, metricActual.position, metricWanted.position)
		} else {
			errorCode, errorMessages := compareMetricData(metricActual.data, metricWanted.data)
			if errorCode != 0 {
				for _, errorMessage := range errorMessages {
					t.Error(errorMessage)
				}
			}
			if errorCode == 2 {
				return
			}
		}
	}
}

package slo

import (
	"fmt"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type slo struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &slo{}, Name: "slo"},
		{F: &slo{}, Name: "sloErrorBudget"},
	}
}

func (f *slo) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"slo": {
			Description: "Returns ratio of points which are in `interval` range and are above/below (`method`) than `value`.\n\nExample:\n\n.. code-block:: none\n\n  &target=slo(some.data.series, \"1hour\", \"above\", 117)",
			Function:    "slo(seriesList, interval, method, value)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "slo",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "interval",
					Required: true,
					Suggestions: types.NewSuggestions(
						"10min",
						"1h",
						"1d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion("above"),
					Name:    "method",
					Options: []string{
						"above",
						"aboveOrEqual",
						"below",
						"belowOrEqual",
					},
					Required: true,
					Type:     types.String,
				},
				{
					Default:  types.NewSuggestion(0.0),
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
			},
		},
		"sloErrorBudget": {
			Description: "Returns rest failure/error budget for this time interval\n\nExample:\n\n.. code-block:: none\n\n  &target=sloErrorBudget(some.data.series, \"1hour\", \"above\", 117, 9999e-4)",
			Group:       "Transform",
			Function:    "sloErrorBudget(seriesList, interval, method, value, objective)",
			Module:      "graphite.render.functions",
			Name:        "sloErrorBudget",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "interval",
					Required: true,
					Suggestions: types.NewSuggestions(
						"10min",
						"1h",
						"1d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion("above"),
					Name:    "method",
					Options: []string{
						"above",
						"aboveOrEqual",
						"below",
						"belowOrEqual",
					},
					Required: true,
					Type:     types.String,
				},
				{
					Default:  types.NewSuggestion(0.0),
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
				{
					Default:  types.NewSuggestion(9999e-4),
					Name:     "objective",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}

func (f *slo) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// alias(aliasByNode(sumSeries(keepLastValue(resources.monitoring.carbon-clickhouse.graphite-cli*.tcp.metricsReceived, 6)), 3), 'total')
	// slo(#A, '30d', 'above', 0)
	var (
		argsExtended, argsWindowed []*types.MetricData
		bucketSize, windowSize     int32
		delta                      int32
		err                        error
	)

	// requested data points' window
	argsWindowed, err = helper.GetSeriesArg(e.Args()[0], from, until, values)
	if len(argsWindowed) == 0 || err != nil {
		return nil, err
	}

	bucketSize, err = e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}

	// there is an opportunity that requested data points' window is smaller than slo interval
	// e.g.: requesting slo(some.data.series, '30days', above, 0) with window of 6 hours
	// this means that we're gonna need 2 sets of data points:
	// - the first one with range [from, until] - for 6 hours
	// - the second one with range [from - delta, until] - for 30 days
	// the result's time range will be 6 hours anyway
	windowSize = until - from
	if bucketSize > windowSize && !(from == 0 && until == 1) {
		delta = bucketSize - windowSize
		argsExtended, err = helper.GetSeriesArg(e.Args()[0], from-delta, until, values)

		if err != nil {
			return nil, err
		}

		if len(argsExtended) != len(argsWindowed) {
			return nil, fmt.Errorf(
				"MetricData quantity differs: there is %d for [%d, %d] and %d for [%d, %d]",
				len(argsExtended), from-delta, until,
				len(argsWindowed), from, until,
			)
		}
	} else {
		argsExtended = argsWindowed
	}

	value, err := e.GetFloatArg(3)
	if err != nil {
		return nil, err
	}

	methodFoo, methodName, err := f.buildMethod(e, 2, value)
	if err != nil {
		return nil, err
	}

	var (
		isErrorBudget bool
		objective     float64
	)

	isErrorBudget = e.Target() == "sloErrorBudget"
	if isErrorBudget {
		objective, err = e.GetFloatArg(4)
		if err != nil {
			return nil, err
		}
	}

	intervalStringValue := e.Args()[1].StringValue()
	results := make([]*types.MetricData, 0, len(argsWindowed))

	for i, argWnd := range argsWindowed {
		var (
			argExt     *types.MetricData
			resultName string
		)

		if isErrorBudget {
			resultName = fmt.Sprintf("sloErrorBudget(%s, %s, %s, %v, %v)", argWnd.Name, intervalStringValue, methodName, value, objective)
		} else {
			resultName = fmt.Sprintf("slo(%s, %s, %s, %v)", argWnd.Name, intervalStringValue, methodName, value)
		}

		// buckets qty is calculated based on requested window
		bucketsQty := helper.GetBucketsQty(argWnd.StartTime, argWnd.StopTime, bucketSize)

		// result for the given series (argWnd)
		r := &types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      resultName,
			Values:    make([]float64, 0, bucketsQty+1),
			IsAbsent:  make([]bool, 0, bucketsQty+1),
			StepTime:  bucketSize,
			StartTime: argWnd.StartTime,
			StopTime:  argWnd.StopTime,
		}}
		// it's ok to place new element to result and modify it later since it's the pointer
		results = append(results, r)

		// if the granularity of series exceeds bucket size then
		// there are not enough data to do the math
		if argWnd.StepTime > bucketSize {
			for i := int32(0); i < bucketsQty; i++ {
				r.IsAbsent = append(r.IsAbsent, true)
				r.Values = append(r.Values, 0.0)
			}
			continue
		}

		// extended data points set will be used for calculating matched items
		argExt = argsExtended[i]

		// calculate SLO using moving window
		qtyMatched := 0 // bucket matched items quantity
		qtyNotNull := 0 // bucket not-null items quantity
		qtyTotal := 0

		timeCurrent := argExt.StartTime
		timeStop := argExt.StopTime
		timeBucketStarts := timeCurrent
		timeBucketEnds := timeCurrent + bucketSize

		// process full buckets
		for i, argValue := range argExt.Values {
			qtyTotal++

			if !argExt.IsAbsent[i] {
				qtyNotNull++
				if methodFoo(argValue) {
					qtyMatched++
				}
			}

			timeCurrent += argExt.StepTime
			if timeCurrent > timeStop {
				break
			}

			if timeCurrent >= timeBucketEnds { // the bucket ends
				newIsAbsent, newValue := f.buildDataPoint(qtyMatched, qtyNotNull)
				if isErrorBudget && !newIsAbsent {
					newValue = (newValue - objective) * float64(bucketSize)
				}

				r.IsAbsent = append(r.IsAbsent, newIsAbsent)
				r.Values = append(r.Values, newValue)

				// init the next bucket
				qtyMatched = 0
				qtyNotNull = 0
				qtyTotal = 0
				timeBucketStarts = timeCurrent
				timeBucketEnds += bucketSize
			}
		}

		// partial bucket might remain
		if qtyTotal > 0 {
			newIsAbsent, newValue := f.buildDataPoint(qtyMatched, qtyNotNull)
			if isErrorBudget && !newIsAbsent {
				newValue = (newValue - objective) * float64(timeCurrent-timeBucketStarts)
			}

			r.IsAbsent = append(r.IsAbsent, newIsAbsent)
			r.Values = append(r.Values, newValue)
		}
	}

	return results, nil
}

func (f *slo) buildDataPoint(bucketQtyMatched, bucketQtyNotNull int) (isAbsent bool, value float64) {
	if bucketQtyNotNull == 0 {
		isAbsent = true
		value = 0.0
	} else {
		isAbsent = false
		value = float64(bucketQtyMatched) / float64(bucketQtyNotNull)
	}
	return
}

func (f *slo) buildMethod(e parser.Expr, argNumber int, value float64) (func(float64) bool, string, error) {
	var methodFoo func(float64) bool = nil

	methodName, err := e.GetStringArg(argNumber)
	if err != nil {
		return nil, methodName, err
	}

	if methodName == "above" {
		methodFoo = func(testedValue float64) bool {
			return testedValue > value
		}
	}

	if methodName == "aboveOrEqual" {
		methodFoo = func(testedValue float64) bool {
			return testedValue >= value
		}
	}

	if methodName == "below" {
		methodFoo = func(testedValue float64) bool {
			return testedValue < value
		}
	}

	if methodName == "belowOrEqual" {
		methodFoo = func(testedValue float64) bool {
			return testedValue <= value
		}
	}

	if methodFoo == nil {
		return nil, methodName, fmt.Errorf("unknown method `%s`", methodName)
	}

	return methodFoo, methodName, nil
}

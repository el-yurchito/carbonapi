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
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}

func (f *slo) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if len(args) == 0 || err != nil {
		return nil, err
	}

	bucketSize, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}

	value, err := e.GetFloatArg(3)
	if err != nil {
		return nil, err
	}

	methodFoo, methodName, err := f.buildMethod(e, 2, value)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {
		name := fmt.Sprintf("slo(%s, %s, %s, %v)", arg.Name, e.Args()[1].StringValue(), methodName, value)

		// align series boundaries and step size according to the interval
		start, stop := arg.StartTime, arg.StopTime
		bucketsQty := helper.GetBuckets(start, stop, bucketSize)

		// result for the given series (arg)
		r := &types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      name,
			Values:    make([]float64, 0, bucketsQty),
			IsAbsent:  make([]bool, 0, bucketsQty),
			StepTime:  bucketSize,
			StartTime: start,
			StopTime:  stop,
		}}
		// it's ok to place new element to result and modify it later since it's the pointer
		results = append(results, r)

		// if the granularity of series exceeds bucket size then
		// there are not enough data to do the math
		if arg.StepTime > bucketSize {
			for i := int32(0); i < bucketsQty; i++ {
				r.IsAbsent = append(r.IsAbsent, true)
				r.Values = append(r.Values, 0.0)
			}
			continue
		}

		// calculate SLO using moving window
		qtyMatched := 0 // bucket matched items quantity
		qtyNotNull := 0 // bucket not-null items quantity
		qtyTotal := 0
		timeBucketEnds := start + bucketSize
		timeCurrent := start

		// process full buckets
		for i, argValue := range arg.Values {
			qtyTotal++

			if !arg.IsAbsent[i] {
				qtyNotNull++
				if methodFoo(argValue) {
					qtyMatched++
				}
			}

			timeCurrent += arg.StepTime
			if timeCurrent >= stop {
				break
			}

			if timeCurrent >= timeBucketEnds { // the bucket ends
				// slo series data points
				newIsAbsent, newValue := f.buildDataPoint(qtyMatched, qtyNotNull)
				r.IsAbsent = append(r.IsAbsent, newIsAbsent)
				r.Values = append(r.Values, newValue)

				// init the next bucket
				qtyMatched = 0
				qtyNotNull = 0
				qtyTotal = 0
				timeBucketEnds += bucketSize
			}
		}

		// partial bucket might remain
		if qtyTotal > 0 {
			newIsAbsent, newValue := f.buildDataPoint(qtyMatched, qtyNotNull)
			r.IsAbsent = append(r.IsAbsent, newIsAbsent)
			r.Values = append(r.Values, newValue)
		}
	}

	return results, nil
}

func (f *slo) buildDataPoint(bucketQtyMatched, bucketQtyNotNull int) (bool, float64) {
	if bucketQtyNotNull == 0 {
		return true, 0.0
	} else {
		return false, float64(bucketQtyMatched) / float64(bucketQtyNotNull)
	}
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

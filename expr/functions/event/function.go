package event

import (
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type event struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &event{}
	functions := []string{"event"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// event(series)
func (f *event) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err == parser.ErrSeriesDoesNotExist {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args)+1)
	switch len(args) {
	case 0:
		const step = 30
		// round `from` and `until` to nearest `step` seconds
		// rounding inside the interval
		if mod := from % step; mod != 0 {
			from += 30 - mod
		}
		if mod := until % step; mod != 0 {
			until -= mod
		}

		pointsQty := int((until-from)/step) + 1
		r := types.MetricData{}

		r.Name = e.ToString()
		r.StartTime = from
		r.StopTime = until
		r.StepTime = step
		r.IsAbsent = make([]bool, pointsQty)
		r.Values = make([]float64, pointsQty)

		results = append(results, &r)

	case 1:
		arg := args[0]
		r := *arg

		r.Name = fmt.Sprintf("event(%s)", arg.Name)
		r.IsAbsent = make([]bool, len(arg.IsAbsent))
		r.Values = make([]float64, len(arg.Values))
		for i, val := range arg.Values {
			if math.IsNaN(val) {
				r.Values[i] = 0
			} else {
				r.Values[i] = val
			}
		}

		results = append(results, &r)

	default:
		return nil, types.ErrWildcardNotAllowed
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *event) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"event": {
			Description: "Like transformNull(series, 0) but also returns zeroes if series has no data.",
			Function:    "event(series)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "event",
			Params: []types.FunctionParam{
				{
					Name:     "series",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}

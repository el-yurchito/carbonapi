package event

import (
	"fmt"
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

func New(configFile string) []interfaces.FunctionMetadata {
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
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	switch len(arg) {
	case 1:
		a := arg[0]
		r := *a
		r.Name = fmt.Sprintf("event(%s)", a.Name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				v = 0
			}
			r.Values[i] = v
		}
		results = append(results, &r)

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

		length := int((until-from)/step) + 1
		r := types.MetricData{}
		r.Name = e.ToString()
		r.Values = make([]float64, length)
		r.IsAbsent = make([]bool, length)
		r.StartTime = from
		r.StopTime = until
		r.StepTime = step
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

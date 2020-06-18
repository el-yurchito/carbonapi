package scaleAfterTimestamp

import (
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type scaleAfterTimestamp struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &scaleAfterTimestamp{},
		Name: "scaleAfterTimestamp",
	}}
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *scaleAfterTimestamp) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"scaleAfterTimestamp": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and multiplies the datapoint\nby the constant provided at each point after the given timestamp.\n\nExample:\n\n.. code-block:: none\n\n  &target=scale(Server.instance01.threads.busy,10)\n  &target=scale(Server.instance*.threads.busy,10)",
			Function:    "scaleAfterTimestamp(seriesList, factor, timestamp)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "scaleAfterTimestamp",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "factor",
					Required: true,
					Type:     types.Float,
				},
				{
					Name:     "timestamp",
					Required: false,
					Type:     types.Integer,
					Default:  types.NewSuggestion(0),
				},
			},
		},
	}
}

// scaleAfterTimestamp(seriesList, factor, timestamp)
func (f *scaleAfterTimestamp) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	scale, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	timestamp, err := e.GetInt32ArgDefault(2, 0)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("scaleAfterTimestamp(%s,%g,%d)", a.Name, scale, timestamp)
		r.IsAbsent = make([]bool, len(a.Values))
		r.Values = make([]float64, len(a.Values))

		currentTimestamp := a.StartTime
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v
			if currentTimestamp >= timestamp {
				r.Values[i] *= scale
			}

			currentTimestamp += a.StepTime
		}
		results = append(results, &r)
	}
	return results, nil
}

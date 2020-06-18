package scale

import (
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type scale struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &scale{}, Name: "scale"},
		{F: &scale{}, Name: "scaleAfterTimestamp"},
	}
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *scale) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"scale": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and multiplies the datapoint\nby the constant provided at each point.\n\nExample:\n\n.. code-block:: none\n\n  &target=scale(Server.instance01.threads.busy,10)\n  &target=scale(Server.instance*.threads.busy,10)",
			Function:    "scale(seriesList, factor)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "scale",
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
			},
		},
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

// scale(seriesList, factor)
// scaleAfterTimestamp(seriesList, factor, timestamp)
func (f *scale) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
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

	results := make([]*types.MetricData, 0, len(args))
	target := e.Target()

	for _, arg := range args {
		r := *arg
		r.Name = f.makeResultName(target, arg, scale, timestamp)
		r.IsAbsent = make([]bool, len(arg.Values))
		r.Values = make([]float64, len(arg.Values))

		currentTimestamp := arg.StartTime
		for i, v := range arg.Values {
			if arg.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v
			if currentTimestamp >= timestamp {
				r.Values[i] *= scale
			}

			currentTimestamp += arg.StepTime
		}
		results = append(results, &r)
	}

	return results, nil
}

func (f *scale) makeResultName(target string, arg *types.MetricData, scale float64, timestamp int32) string {
	if target == "scale" {
		return fmt.Sprintf("%s(%s,%g)", target, arg.Name, scale)
	} else {
		return fmt.Sprintf("%s(%s,%g,%d)", target, arg.Name, scale, timestamp)
	}
}

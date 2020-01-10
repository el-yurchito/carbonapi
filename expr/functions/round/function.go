package round

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type round struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &round{}
	functions := []string{"round"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// round(seriesList,precision)
func (f *round) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var withPrecision bool
	precision, err := e.GetIntArg(1)
	switch err {
	case nil:
		withPrecision = true
	case parser.ErrMissingArgument:
		// precision is already 0
		withPrecision = false
	default:
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		if withPrecision {
			r.Name = fmt.Sprintf("round(%s,%d)", a.Name, precision)
		} else {
			r.Name = fmt.Sprintf("round(%s)", a.Name)
		}
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = doRound(v, precision)
		}
		results = append(results, &r)
	}
	return results, nil
}

func doRound(x float64, precision int) float64 {
	roundTo := math.Pow10(precision)
	return math.RoundToEven(x*roundTo) / roundTo
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *round) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"round": {
			Description: "Takes one metric or a wildcard seriesList optionally followed by a precision, and rounds each\ndatapoint to the specified precision.\n\nExample:\n\n.. code-block:: none\n\n  &target=round(Server.instance01.threads.busy)\n  &target=round(Server.instance01.threads.busy,2)",
			Function:    "round(seriesList, precision)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "round",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "precision",
					Required: false,
					Type:     types.Integer,
				},
			},
		},
	}
}

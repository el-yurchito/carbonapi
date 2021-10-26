package diffSeries

import (
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type diffSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &diffSeries{}
	functions := []string{"diffSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// diffSeries(*seriesLists)
func (f *diffSeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (result []*types.MetricData, err error) {
	minuends, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	subtrahends, err := helper.GetSeriesArgs(e.Args()[1:], from, until, values)
	if err != nil {
		return nil, err
	}

	var (
		arg2Name string
		subsQty  int
	)
	if subsQty = len(subtrahends); subsQty == 0 {
		arg2Name = "nil"
	} else if subsQty == 1 {
		arg2Name = subtrahends[0].Name
	} else {
		arg2Name = helper.RemoveEmptySeriesFromName(subtrahends)
	}

	result = make([]*types.MetricData, len(minuends))
	for i, minuend := range minuends {
		r := *minuend
		r.Name = fmt.Sprintf("diffSeries(%s,%s)", minuend.Name, arg2Name)
		r.IsAbsent = make([]bool, len(minuend.Values))
		r.Values = make([]float64, len(minuend.Values))

		for j := range minuend.Values {
			if minuend.IsAbsent[j] {
				r.IsAbsent[j] = true
				continue
			}

			var subsSum float64
			for _, sub := range subtrahends {
				jSub := (int32(j) * minuend.StepTime) / sub.StepTime
				if sub.IsAbsent[jSub] {
					continue
				}
				subsSum += sub.Values[jSub]
			}
			r.Values[j] = minuend.Values[j] - subsSum
		}
		result[i] = &r
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *diffSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"diffSeries": {
			Description: "Subtracts series 2 through n from series 1.\n\nExample:\n\n.. code-block:: none\n\n  &target=diffSeries(service.connections.total,service.connections.failed)\n\nTo diff a series and a constant, one should use offset instead of (or in\naddition to) diffSeries\n\nExample:\n\n.. code-block:: none\n\n  &target=offset(service.connections.total,-5)\n\n  &target=offset(diffSeries(service.connections.total,service.connections.failed),-4)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``diff``.",
			Function:    "diffSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "diffSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}

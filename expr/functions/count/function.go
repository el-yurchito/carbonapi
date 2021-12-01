package count

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type count struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{Name: "count", F: &count{}}}
}

// count(seriesList)
func (f *count) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	e.SetTarget("count")
	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		return float64(len(values))
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *count) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"count": {
			Description: "Draws a horizontal line representing the number of nodes found in the seriesList.\n\n.. code-block:: none\n\n  &target=countSeries(carbon.agents.*.*)",
			Function:    "count(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "count",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Type:     types.SeriesList,
				},
			},
		},
	}
}

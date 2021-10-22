package groupWithSql

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type groupWithSql struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{Name: "groupWithSql", F: &groupWithSql{}}}
}

func (f *groupWithSql) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var results []*types.MetricData
	key := parser.MetricRequest{Metric: e.ToString(), From: from, Until: until}
	data, ok := values[key]
	if !ok {
		return results, nil
	}
	return data, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *groupWithSql) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"groupWithSql": {
			Description: `
Maps a callback to subgroups within as defined by multiple tags

.. code-block:: none

groupWithSql('sum', 'name=cpu', 'cluster=romeo')

Would return multiple series which are each the result of applying the "sum" function
to groups joined on the specified tags .
`,
			Function: "groupWithSql(callback, *tagExpressions)",
			Group:    "Special",
			Module:   "graphite.render.functions",
			Name:     "groupWithSql",
			Params: []types.FunctionParam{
				{
					Name:     "callback",
					Options:  helper.AvailableSummarizers,
					Required: true,
					Type:     types.AggFunc,
				},
				{
					Name:     "tagExpressions",
					Required: true,
					Type:     types.String,
					Multiple: true,
				},
			},
		},
	}
}

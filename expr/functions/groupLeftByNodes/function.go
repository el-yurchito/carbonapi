package groupLeftByNodes

import (
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type groupLeftByNodes struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &groupLeftByNodes{}, Name: "groupLeftByNodes"},
	}
}

func (f *groupLeftByNodes) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"groupLeftByNodes": {
			Description: `Performs operation similar to "group_left" in prometheus. Used on non-tagged metrics only.
Each series from "seriesList1" is matched with each series from "seriesList2" based on values at specified nodes.
Name of matched series from the second list is appended to respective names from the first list
`,
			Function: "groupLeftByNodes(seriesList1, seriesList2, 0, 1, -2, ...)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "groupLeftByNodes",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList1",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesList2",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Required: true,
					Name:     "nodes",
					Type:     types.Node,
				},
			},
		},
	}
}

func (f *groupLeftByNodes) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesList1, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	seriesList2, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	nodes, err := e.GetIntArgs(2)
	if err == parser.ErrMissingArgument {
		return nil, fmt.Errorf("at least one node must be specified")
	}

	grouped2 := groupByKey(seriesList2, nodes)
	results = make([]*types.MetricData, 0, len(seriesList1))
	for _, series1 := range seriesList1 {
		key := safeKeyByNodes(series1.Name, nodes)
		list2, ok := grouped2[key]
		if !ok {
			continue
		}

		for _, data2 := range list2 {
			data1 := *series1
			data1.Name = fmt.Sprintf("%s.%s", data1.Name, data2.Name)
			results = append(results, &data1)
		}
	}
	return results, nil
}

func groupByKey(list []*types.MetricData, keyNodes []int) map[string][]*types.MetricData {
	result := make(map[string][]*types.MetricData, len(list))
	for _, data := range list {
		key := safeKeyByNodes(data.Name, keyNodes)
		result[key] = append(result[key], data)
	}
	return result
}

const sep = "."

func safeKeyByNodes(metric string, keyNodes []int) string {
	parts := strings.Split(metric, sep)
	partsQty := len(parts)

	result := strings.Builder{}
	for _, node := range keyNodes {
		if node >= 0 && node < partsQty {
			result.WriteString(parts[node])
		} else if node < 0 && node >= -partsQty {
			result.WriteString(parts[partsQty+node])
		}
		result.WriteString(sep)
	}
	return result.String()
}

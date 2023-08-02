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
Name of matched series from the second list is appended to respective name from the first list.
Following options for "valuesFrom" parameter are available:
  * "default" - values in result will be taken from the first series. If you need values from the second series,
	just swap "seriesList1" and "seriesList2" arguments.
  * "common" - values in result also will be taken from the first series, but for each absent value from the
	second series, respective value from the first series will be absent as well. For example, two metrics have matched:
	"aaa.bbb.ccc" with values [1, 2, 3, 4, 5] from the first list and "aaa.bbb.ddd" with values [5, null, 3, 2, null] 
	from the second list, therefore result metric will be "aaa.bbb.ccc.aaa.bbb.ddd" with values [1, null, 3, 4, null].
`,
			Function: "groupLeftByNodes(seriesList1, seriesList2, 'default', 0, 1, -2, ...)",
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
					Name:     "valuesFrom",
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion(valuesDefault),
					Options:  []string{valuesDefault, valuesCommon},
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

const (
	valuesDefault = "default"
	valuesCommon  = "common"
)

func (f *groupLeftByNodes) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesList1, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	seriesList2, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	valuesFrom, err := e.GetStringNamedOrPosArgDefault("valuesFrom", 2, valuesDefault)
	if err != nil {
		return nil, err
	}
	valuesFrom = strings.ToLower(valuesFrom)
	if valuesFrom != valuesDefault && valuesFrom != valuesCommon {
		return nil, fmt.Errorf(
			"unknown %q value: only %q and %q are allowed",
			valuesFrom, valuesDefault, valuesCommon,
		)
	}

	nodes, err := e.GetIntArgs(3)
	if err == parser.ErrMissingArgument {
		return nil, fmt.Errorf("at least one node must be specified")
	}

	var mergeMetricData func(series1, series2 *types.MetricData) *types.MetricData
	switch strings.ToLower(valuesFrom) {
	case valuesDefault:
		mergeMetricData = mergeMetricDataDefault
	case valuesCommon:
		mergeMetricData = mergeMetricDataCommon
	default:
		return nil, fmt.Errorf(
			"unknown valuesFrom value %q: only %q and %q are allowed",
			valuesFrom, valuesDefault, valuesCommon,
		)
	}

	grouped2 := groupByKey(seriesList2, nodes)
	results = make([]*types.MetricData, 0, len(seriesList1))
	for _, series1 := range seriesList1 {
		key := safeKeyByNodes(series1.Name, nodes)
		for _, series2 := range grouped2[key] {
			results = append(results, mergeMetricData(series1, series2))
		}
	}
	return results, nil
}

func copyMetricData(md *types.MetricData) *types.MetricData {
	data := *md
	result := &data

	result.IsAbsent = make([]bool, len(md.IsAbsent))
	copy(result.IsAbsent, md.IsAbsent)

	result.Values = make([]float64, len(md.Values))
	copy(result.Values, md.Values)

	return result
}

func groupByKey(list []*types.MetricData, keyNodes []int) map[string][]*types.MetricData {
	result := make(map[string][]*types.MetricData, len(list))
	for _, data := range list {
		key := safeKeyByNodes(data.Name, keyNodes)
		result[key] = append(result[key], data)
	}
	return result
}

func mergeMetricDataCommon(series1, series2 *types.MetricData) *types.MetricData {
	result := copyMetricData(series1)
	result.Name = fmt.Sprintf("%s.%s", series1.Name, series2.Name)

	for i := range result.Values {
		if i >= len(result.IsAbsent) || i >= len(series2.IsAbsent) {
			continue
		}

		if !result.IsAbsent[i] && series2.IsAbsent[i] {
			result.Values[i] = 0
			result.IsAbsent[i] = true
		}
	}

	return result
}

func mergeMetricDataDefault(series1, series2 *types.MetricData) *types.MetricData {
	result := copyMetricData(series1)
	result.Name = fmt.Sprintf("%s.%s", series1.Name, series2.Name)
	return result
}

const sep = "."

func safeKeyByNodes(metric string, keyNodes []int) string {
	parts := strings.Split(metric, sep)
	partsQty := len(parts)

	result := strings.Builder{}
	for i, node := range keyNodes {
		if node >= 0 && node < partsQty {
			result.WriteString(parts[node])
		} else if node < 0 && node >= -partsQty {
			result.WriteString(parts[partsQty+node])
		}

		if i != len(keyNodes)-1 {
			result.WriteString(sep)
		}
	}
	return result.String()
}

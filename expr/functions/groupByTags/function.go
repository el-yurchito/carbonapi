package groupByTags

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type groupByTags struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &groupByTags{},
		Name: "groupByTags",
	}}
}

// seriesByTag("name=cpu")|groupByTags("average","dc","os")
func (f *groupByTags) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	callback, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	tags, err := e.GetStringArgs(2)
	if err != nil {
		return nil, err
	}
	sort.Strings(tags)

	groupedVales := make(map[string][]*types.MetricData)
	resultSeriesList := make([]*types.MetricData, 0, len(seriesList))

	// group metric values by specified tags
	for _, series := range seriesList {
		keyParts := make([]string, len(tags))
		metricTags := helper.ExtractTags(series.Name)

		// constructing the key: it is string that looks like ".val1.val2.val3"
		// where val1, val2 and val3 - tags' values
		for i, tag := range tags {
			metricTagValue := metricTags[tag]
			keyParts[i] = metricTagValue
		}

		key := strings.Join(keyParts, types.MetricPathSep)
		groupedVales[key] = append(groupedVales[key], series)
	}

	for key, values := range groupedVales {
		expr := fmt.Sprintf("%s(stub)", callback)
		stubExpr, _, err := parser.ParseExpr(expr) // create a stub context to evaluate the callback in
		if err != nil {
			return nil, err
		}

		stubValues := map[parser.MetricRequest][]*types.MetricData{
			parser.MetricRequest{
				Metric: "stub",
				From:   from,
				Until:  until,
			}: values,
		}
		exprEvaluated, _ := f.Evaluator.EvalExpr(stubExpr, from, until, stubValues)
		if exprEvaluated != nil {
			exprEvaluated[0].Name = fmt.Sprintf("%s", key) // copy
			resultSeriesList = append(resultSeriesList, exprEvaluated...)
		}
	}

	return resultSeriesList, nil
}

func (f *groupByTags) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"groupByTags": {
			Description: "Takes a serieslist and maps a callback to subgroups within as defined by multiple tags\n\n.. code-block:: none\n\n  &target=seriesByTag(\"name=cpu\")|groupByTags(\"average\",\"dc\")\n\nWould return multiple series which are each the result of applying the \"averageSeries\" function\nto groups joined on the specified tags resulting in a list of targets like\n\n.. code-block :: none\n\n  averageSeries(seriesByTag(\"name=cpu\",\"dc=dc1\")),averageSeries(seriesByTag(\"name=cpu\",\"dc=dc2\")),...\n\nThis function can be used with all aggregation functions supported by\n:py:func:`aggregate <aggregate>`: ``average``, ``median``, ``sum``, ``min``, ``max``, ``diff``,\n``stddev``, ``range`` & ``multiply``.",
			Function:    "groupByTags(seriesList, callback, *tags)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "groupByTags",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "callback",
					Options:  helper.AvailableSummarizers,
					Required: true,
					Type:     types.AggFunc,
				},
				{
					Name:     "tags",
					Required: true,
					Multiple: true,
					Type:     types.Tag,
				},
			},
		},
	}
}

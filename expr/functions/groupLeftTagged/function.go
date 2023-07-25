package groupLeftTagged

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type groupLeftTagged struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &groupLeftTagged{}, Name: "groupLeftTagged"},
	}
}

func (f *groupLeftTagged) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"groupLeftTagged": {
			Description: `Performs operation similar to "group_left" in prometheus. Used on tagged metrics only.
Each series from "seriesList1" is matched with each series from "seriesList2" based on values of specified tags.
Matched series from the first list receives all tags from series from the second list.
`,
			Function: "groupLeftTagged(seriesList1, seriesList2, 'tag1', 'tag2', ...)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "groupLeftTagged",
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
					Name:     "tags",
					Type:     types.Tag,
				},
			},
		},
	}
}

func (f *groupLeftTagged) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesList1, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	seriesList2, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	tags, err := e.GetStringArgs(2)
	if err == parser.ErrMissingArgument {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		tags = []string{nameTag}
	}

	parsed1, parsed2 := parseSeriesList(seriesList1, tags), parseSeriesList(seriesList2, tags)
	grouped2 := groupByKey(parsed2)
	results = make([]*types.MetricData, 0, len(seriesList1))

	for _, series1 := range parsed1 {
		for _, md2 := range grouped2[series1.key] {
			data1 := *series1.data
			data1.Name = series1.parsed.merge(md2).String()
			results = append(results, &data1)
		}
	}

	return results, nil
}

type metricDataWithKey struct {
	data   *types.MetricData
	parsed *taggedMetric
	key    string
}

func parseSeriesList(list []*types.MetricData, tags []string) []metricDataWithKey {
	result := make([]metricDataWithKey, 0, len(list))
	for _, data := range list {
		parsed := parseTaggedMetric(data.Name)
		result = append(result, metricDataWithKey{
			data:   data,
			parsed: parsed,
			key:    parsed.key(tags),
		})
	}
	return result
}

func groupByKey(parsedList []metricDataWithKey) map[string][]*taggedMetric {
	result := make(map[string][]*taggedMetric, len(parsedList))
	for _, parsed := range parsedList {
		result[parsed.key] = append(result[parsed.key], parsed.parsed)
	}
	return result
}

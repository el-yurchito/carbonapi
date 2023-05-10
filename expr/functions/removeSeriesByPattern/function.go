package removeSeriesByPattern

import (
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func GetOrder() interfaces.Order {
	return interfaces.Any
}

type removeSeriesByPattern struct {
	interfaces.FunctionBase
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		Name: "removeSeriesByPattern",
		F:    &removeSeriesByPattern{},
	}}
}

func (f *removeSeriesByPattern) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	mainList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	patternList, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	// transform patterns list to set
	patterns := make(map[string]bool, len(patternList))
	for _, md := range patternList {
		patterns[md.Name] = true
	}

	// process main series list: exclude all series that matched any pattern
	result := make([]*types.MetricData, 0, len(mainList))
	for _, md := range mainList {
		parts := strings.Split(md.Name, ".")
		matched := false

		// compare series parts with pattern
		// use only strict match, no regexps
		for _, part := range parts {
			if patterns[part] {
				matched = true
				break
			}
		}

		if !matched {
			result = append(result, md)
		}
	}

	return result, nil
}

func (f *removeSeriesByPattern) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"removeSeriesByPattern": {
			Description: "Takes a metric or a wildcard seriesList, followed by a metric or wildcard patternSeriesList.\nExcludes all series from main list that matched any pattern.",
			Function:    "removeSeriesByPattern(seriesList, patternSeriesList)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "removeSeriesByPattern",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "patternSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}

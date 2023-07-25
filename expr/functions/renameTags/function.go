package renameTags

import (
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type renameTags struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &renameTags{}, Name: "renameTags"},
	}
}

func (f *renameTags) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"renameTags": {
			Description: `Renames tags for each series of "seriesList"`,
			Function:    "renameTags(seriesList, 'oldTagName1', 'newTagName1', 'oldTagName2', 'newTagName2', ...)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "renameTags",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "replacements",
					Type:     types.Tag,
				},
			},
		},
	}
}

func (f *renameTags) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	replacements, err := e.GetStringArgs(1)
	if err == parser.ErrMissingArgument {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if len(replacements)%2 != 0 {
		return nil, fmt.Errorf("number of replacements must be even")
	}

	results = make([]*types.MetricData, 0, len(seriesList))
	replacementsMap := makeReplacementsMap(replacements)
	for _, series := range seriesList {
		r := *series
		r.Name = runReplacements(r.Name, replacementsMap)
		results = append(results, &r)
	}
	return results, nil
}

func makeReplacementsMap(replacements []string) map[string]string {
	result := make(map[string]string, len(replacements)/2)
	for i := 0; i < len(replacements); i += 2 {
		result[replacements[i]] = replacements[i+1]
	}
	return result
}

func runReplacements(name string, replacementsMap map[string]string) string {
	tags := helper.ExtractTags(name)
	for oldTag, newTag := range replacementsMap {
		if oldTagVal, ok := tags[oldTag]; ok {
			oldStr := fmt.Sprintf("%s=%s", oldTag, oldTagVal)
			newStr := fmt.Sprintf("%s=%s", newTag, oldTagVal)
			name = strings.ReplaceAll(name, oldStr, newStr)
		}
	}
	return name
}

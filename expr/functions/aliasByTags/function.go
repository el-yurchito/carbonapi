package aliasByTags

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const NAME = "name"

type aliasByTags struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aliasByTags{}
	for _, n := range []string{"aliasByTags"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

//
func extractTagValue(tagValue string) string {
	buffer := bytes.Buffer{}
	tagValueEnds := len(tagValue)

	for i := 0; i < len(tagValue); i++ {
		if tagValue[i] == '(' || tagValue[i] == ')' || tagValue[i] == ',' {
			tagValueEnds = i
			break
		}
	}

	buffer.WriteString(tagValue[:tagValueEnds])
	return buffer.String()
}

// metricToTagMap splits metric with tags to pure metric and tag:value map separately
func metricToTagMap(metric string) map[string]string {
	removals := make([]string, 100) // tokens to be removed
	result := make(map[string]string)

	for _, metricPart := range strings.Split(metric, ";") {
		tagsParts := strings.SplitN(metricPart, "=", 2)
		if len(tagsParts) == 2 {
			tagName := tagsParts[0]
			tagValue := extractTagValue(tagsParts[1])

			removals = append(removals, fmt.Sprintf("%s=%s", tagName, tagValue))
			result[tagName] = tagValue
		}
	}

	// semicolons will be removed as well
	removals = append(removals, ";")
	metricCleaned := fmt.Sprintf("%s", metric)
	for _, removal := range removals {
		metricCleaned = strings.Replace(metricCleaned, removal, "", -1)
	}

	result[NAME] = metricCleaned
	return result
}

func (f *aliasByTags) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		fmt.Println("getSeriesArg missing argument")
		return nil, err
	}

	tags, err := e.GetNodeOrTagArgs(1)
	if err != nil {
		fmt.Println("GetNodeOrTagArgs missing argument")
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		var matched []string
		metricTags := metricToTagMap(a.Name)
		fmt.Printf("!!!!!!\na.Name = %s\nmetricTags = %#v\n", a.Name, metricTags)
		nodes := strings.Split(metricTags["name"], ".")
		for _, tag := range tags {
			if tag.IsTag {
				tagStr := tag.Value.(string)
				matched = append(matched, metricTags[tagStr])
			} else {
				f := tag.Value.(int)
				if f < 0 {
					f += len(nodes)
				}
				if f >= len(nodes) || f < 0 {
					continue
				}
				matched = append(matched, nodes[f])
			}
		}
		r := *a
		if len(matched) > 0 {
			r.Name = strings.Join(matched, ".")
		}
		r.Name = strings.Split(r.Name, ",")[0]
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aliasByTags) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByTags": {
			Description: "Takes a seriesList and applies an alias derived from one or more tags",
			Function:    "aliasByTags(seriesList, *tags)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByTags",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "tags",
					Required: true,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}

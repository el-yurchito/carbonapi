package aliasByBase64

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"encoding/base64"
	"strings"
)

type aliasByBase64 struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aliasByBase64{}
	for _, n := range []string{"aliasByBase64"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasByBase64) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	field, err := e.GetIntArg(1)
	withoutFieldArg := err != nil

	var results []*types.MetricData

	for _, a := range args {
		r := *a
		if withoutFieldArg {
			decoded, err := base64.StdEncoding.DecodeString(r.Name)
			if err == nil {
				r.Name = string(decoded)
			}
		} else {
			metric := helper.ExtractMetric(r.Name)
			var name []string
			for i, n := range strings.Split(metric, ".") {
				if i == field {
					decoded, err := base64.StdEncoding.DecodeString(n)
					if err == nil {
						n = string(decoded)
					}
				}
				name = append(name, n)
			}
			r.Name = strings.Join(name, ".")
		}

		results = append(results, &r)
	}

	return results, nil
}

func (f *aliasByBase64) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByBase64": {
			Description: "Takes a seriesList and decodes its name with base64\n\n.. code-block:: none\n\n  &target=aliasByMetric(carbon.agents.graphite.creates)",
			Function:    "aliasByBase64(seriesList)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByBase64",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "nodeNum",
					Required: false,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}

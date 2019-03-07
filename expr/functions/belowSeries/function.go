package belowSeries

import (
	"regexp"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type belowSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &belowSeries{}
	functions := []string{"belowSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *belowSeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	min, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	rename := true
	search, err := e.GetStringArg(2)
	if err != nil {
		rename = false
	}

	replace, err := e.GetStringArg(3)
	if err != nil {
		rename = false
	}

	var rre *regexp.Regexp
	if rename {
		rre, err = regexp.Compile(search)
		if err != nil {
			return nil, err
		}
	}

	var results []*types.MetricData
	for _, a := range args {
		if helper.MinValue2(a.Values) < min {
			r := *a
			if rename {
				r.Name = rre.ReplaceAllString(r.Name, replace)
			}
			results = append(results, &r)
		}
	}

	return results, nil
}

func (f *belowSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"belowSeries": {
			Name:        "belowSeries",
			Description: "Takes a seriesList and compares the minimum of each series against the given value. If the series minimum is less than value, the regular expression search and replace is applied against the series name to plot a related metric e.g. given useSeriesBelow, the response time metric will be plotted only when the minimum value of the corresponding request/s metric is < 10",
			Function:    "belowSeries(seriesList, value, search, replace)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "value",
					Required: true,
					Type:     types.Integer,
				},
				{
					Name: "search",
					Type: types.String,
				},
				{
					Name: "replace",
					Type: types.String,
				},
			},
		},
	}
}

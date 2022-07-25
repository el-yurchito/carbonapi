package removeEmptySeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type removeEmptySeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &removeEmptySeries{}
	functions := []string{"removeEmptySeries", "removeZeroSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// removeEmptySeries(seriesLists, n), removeZeroSeries(seriesLists, n)
func (f *removeEmptySeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	factor, err := e.GetFloatArgDefault(1, 0.0)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {
		notNull := 0
		for i, v := range arg.Values {
			if !math.IsNaN(v) && !arg.IsAbsent[i] {
				switch e.Target() {
				case "removeEmptySeries":
					notNull++
				case "removeZeroSeries":
					if v != 0 {
						notNull++
					}
				}
			}
		}
		if notNull != 0 && float64(notNull)/float64(len(arg.Values)) >= factor {
			results = append(results, arg)
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *removeEmptySeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"removeEmptySeries": {
			Description: "Takes one metric or a wildcard seriesList.\nOut of all metrics passed, draws only the metrics with not empty data\n\nExample:\n\n.. code-block:: none\n\n  &target=removeEmptySeries(server*.instance*.threads.busy)\n\nDraws only live servers with not empty data.\n\n`xFilesFactor` follows the same semantics as in Whisper storage schemas.  Setting it to 0 (the\ndefault) means that only a single value in the series needs to be non-null for it to be\nconsidered non-empty, setting it to 1 means that all values in the series must be non-null.\nA setting of 0.5 means that at least half the values in the series must be non-null.",
			Function:    "removeEmptySeries(seriesList, xFilesFactor=None)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "removeEmptySeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "xFilesFactor",
					Required: true,
					Type:     types.Float,
				},
			},
		},
		"removeZeroSeries": {
			Description: "Takes one metric or a wildcard seriesList.\nOut of all metrics passed, draws only the metrics with not ZERO data\n\nExample:\n\n.. code-block:: none\n\n  &target=removeZeroSeries(server*.instance*.threads.busy)\n\nDraws only live servers with not empty data.\n\n`xFilesFactor` follows the same semantics as in Whisper storage schemas.  Setting it to 0 (the\ndefault) means that only a single value in the series needs to be non-null for it to be\nconsidered non-empty, setting it to 1 means that all values in the series must be non-null.\nA setting of 0.5 means that at least half the values in the series must be non-null.",
			Function:    "removeZeroSeries(seriesList, xFilesFactor=None)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions.custom",
			Name:        "removeZeroSeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "xFilesFactor",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}

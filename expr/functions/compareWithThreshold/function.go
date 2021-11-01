package compareWithThreshold

import (
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type compareWithThreshold struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &compareWithThreshold{}
	for _, n := range []string{"compareWithThreshold"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *compareWithThreshold) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	thresholds, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	defaultThreshold, err := e.GetFloatArg(2)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		// try to find a threshold with matching name
		var threshold *types.MetricData

		nameEndsAt := strings.IndexByte(a.Name, ';')
		if nameEndsAt == -1 {
			continue
		}
		tags := a.Name[nameEndsAt:]
		for _, th := range thresholds {
			// check that `th.Name` matches `someStringWithoutSemicolon;tags`
			if strings.HasSuffix(th.Name, tags) && strings.IndexByte(strings.TrimSuffix(th.Name, tags), ';') == -1 {
				threshold = th
				break
			}
		}

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				continue
			}

			if threshold == nil {
				if v >= defaultThreshold {
					results = append(results, a)
					break
				}

			} else {
				iThreshold := (int32(i) * a.StepTime) / threshold.StepTime
				if threshold.IsAbsent[iThreshold] {
					continue
				}

				if v >= threshold.Values[iThreshold] {
					results = append(results, a)
					break
				}

			}

		}
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *compareWithThreshold) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"compareWithThreshold": {
			Description: "Compares a list of tagged metrics with a list of tagged threshold metrics.",
			Function:    "compareWithThreshold(seriesList, thresholdSeriesList, defaultThreshold)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "compareWithThreshold",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "thresholdSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "default",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}

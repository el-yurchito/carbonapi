package heatMap

import (
	"fmt"
	"sort"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type heatMap struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &heatMap{},
		Name: "heatMap",
	}}
}

func (f *heatMap) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"heatMap": {
			Description: "Assume seriesList has values N values in total: (a[1], a[2], ..., a[N]). Then heatMap(seriesList) has N-1 values in total: (a[2] - a[1], a[3] - a[2], ..., a[N] - a[N-1]).",
			Function:    "heatMap(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "heatMap",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Proxied: true,
		},
	}
}

func (f *heatMap) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (resultData []*types.MetricData, resultError error) {
	series, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	series = f.sortMetricData(series)
	seriesQty := len(series)
	result := make([]*types.MetricData, 0, seriesQty-1)

	for i := 1; i < seriesQty; i++ {
		curr, prev := series[i], series[i-1]
		if err := f.validateNeighbourSeries(curr, prev); err != nil {
			return nil, err
		}

		pointsQty := len(curr.Values)
		r := &types.MetricData{FetchResponse: types.FetchResponse{
			Name:      fmt.Sprintf("heatMap(%s,%s)", curr.Name, prev.Name),
			IsAbsent:  make([]bool, pointsQty),
			Values:    make([]float64, pointsQty),
			StartTime: curr.StartTime,
			StopTime:  curr.StopTime,
			StepTime:  curr.StepTime,
		}}

		for j := 0; j < pointsQty; j++ {
			r.IsAbsent[j] = curr.IsAbsent[j] || prev.IsAbsent[j]
			if !r.IsAbsent[j] {
				r.Values[j] = curr.Values[j] - prev.Values[j]
			}
		}

		result = append(result, r)
	}

	return result, nil
}

// sortMetricData returns *types.MetricData list sorted by sum of the first values
func (f *heatMap) sortMetricData(list []*types.MetricData) []*types.MetricData {
	// take 5 first not null values
	const points = 5

	// mate series with its weight (sum of first values)
	type metricDataWeighted struct {
		data   *types.MetricData
		weight float64
	}

	seriesQty := len(list)
	if seriesQty < 2 {
		return list
	}

	listWeighted := make([]metricDataWeighted, seriesQty)
	for j := 0; j < seriesQty; j++ {
		listWeighted[j].data = list[j]
	}

	pointsFound := 0
	valuesQty := len(list[0].Values)

	for i := 0; i < valuesQty && pointsFound < points; i++ {
		// make sure that each series has current point not null
		absent := false
		for j := 0; j < seriesQty && !absent; j++ {
			absent = list[j].IsAbsent[i]
		}
		if absent {
			continue
		}

		// accumulate sum of first not-null values
		for j := 0; j < seriesQty; j++ {
			listWeighted[j].weight += list[j].Values[i]
		}
		pointsFound++
	}

	// sort series by its weight
	if pointsFound > 0 {
		sort.SliceStable(listWeighted, func(i, j int) bool {
			return listWeighted[i].weight < listWeighted[j].weight
		})
		for j := 0; j < seriesQty; j++ {
			list[j] = listWeighted[j].data
		}
	}

	return list
}

func (f *heatMap) validateNeighbourSeries(s1, s2 *types.MetricData) error {
	if s1.StartTime != s2.StartTime {
		return fmt.Errorf("StartTime differs: %d!=%d", s1.StartTime, s2.StartTime)
	}
	if s1.StopTime != s2.StopTime {
		return fmt.Errorf("StartTime differs: %d!=%d", s1.StopTime, s2.StopTime)
	}
	if s1.StepTime != s2.StepTime {
		return fmt.Errorf("StartTime differs: %d!=%d", s1.StepTime, s2.StepTime)
	}
	if len(s1.Values) != len(s2.Values) {
		return fmt.Errorf("values quantity differs: %d!=%d", len(s1.Values), len(s2.Values))
	}
	for _, s := range []*types.MetricData{s1, s2} {
		if len(s.IsAbsent) != len(s.Values) {
			return fmt.Errorf("values and isAbsent quantities differ for %s: %d!=%d", s.Name, len(s.Values), len(s.IsAbsent))
		}
	}
	return nil
}

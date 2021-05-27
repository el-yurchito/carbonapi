package anomaly

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const anomalyPrefix = "resources.monitoring.anomaly_detector."

type anomaly struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &anomaly{}
	functions := []string{"anomaly"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *anomaly) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	joinType, err := e.GetStringNamedOrPosArgDefault("type", 1, "all")
	if err != nil {
		return nil, err
	}
	threshold, err := e.GetFloatNamedOrPosArgDefault("threshold", 2, math.NaN())
	if err != nil {
		return nil, err
	}

	offs, err := e.GetIntervalArgDefault(3, 1, -1)
	if err != nil {
		return nil, err
	}

	// extract anomaly metrics
	anomalyMap := make(map[string]*types.MetricData)
	for _, mr := range e.Args()[0].Metrics() {
		metric := anomalyPrefix + mr.Metric
		for _, data := range values[parser.MetricRequest{Metric: metric, From: from, Until: until}] {
			if offs > 0 {
				offPoints := (data.StopTime - offs - data.StartTime) / data.StepTime
				if offPoints < 0 {
					offPoints = 0
				}

				exclude := true
				for _, v := range data.IsAbsent[offPoints:] {
					if !v {
						exclude = false
						break
					}
				}
				if exclude {
					continue
				}
			}

			name := strings.TrimPrefix(data.Name, anomalyPrefix)
			data.Name = fmt.Sprintf("[anomaly] %s", name)
			anomalyMap[name] = data
		}
	}

	var results []*types.MetricData
	for _, a := range args {
		exclude := false
		if !math.IsNaN(threshold) {
			exclude = true
			for i, v := range a.Values {
				if !a.IsAbsent[i] && v > threshold {
					exclude = false
					break
				}
			}
		}
		if exclude {
			continue
		}
		anomaly, hasAnomaly := anomalyMap[a.Name]
		// include all metrics & anomalies
		if joinType == "all" {
			results = append(results, a)
			if hasAnomaly {
				results = append(results, anomaly)
			}
		} else if joinType == "with_anomalies_only" && hasAnomaly {
			results = append(results, a)
			results = append(results, anomaly)
		} else if joinType == "only_anomalies" && hasAnomaly {
			results = append(results, anomaly)
		}
	}
	return results, nil
}

const descr = `Принимает на вход метрику (или массив метрик), выводит помимо метрик их аномальные точки.
Параметр type, принимает значения:
all - выводить все метрики и аномалии (дефолт)
with_anomalies_only - выводить только метрики с аномалиями + аномалии
only_anomalies - выводить только аномалии
`

func (f *anomaly) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"anomaly": {
			Description: descr,
			Function:    "anomaly(seriesList, type='all')",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "anomaly",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "type",
					Options: []string{
						"all",
						"with_anomalies_only",
						"only_anomalies",
					},
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion("all"),
				},
				// TODO add offset & threshold
			},
		},
	}
}

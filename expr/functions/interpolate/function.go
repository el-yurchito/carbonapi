package interpolate

import (
	"fmt"
	"math"
	"runtime/debug"

	"github.com/lomik/zapwriter"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type interpolate struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &interpolate{},
		Name: "interpolate",
	}}
}

func (f *interpolate) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"interpolate": {
			Description: "Takes one metric or a wildcard seriesList, and optionally a limit to the number of 'None' values to skip over." +
				"\nContinues the line with the last received value when gaps ('None' values) appear in your data, rather than breaking your line." +
				"\n\n.. code-block:: none\n\n  &target=interpolate(Server01.connections.handled)\n  &target=interpolate(Server01.connections.handled, 10)",
			Function: "interpolate(seriesList, limit)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "interpolate",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "limit",
					Required: false,
					Type:     types.Float,
				},
			},
			Proxied: true,
		},
	}
}

func (f *interpolate) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (resultData []*types.MetricData, resultError error) {
	sugaredLogger := zapwriter.Logger("functionDo").With(zap.String("function", "interpolate")).Sugar()
	defer func() {
		if r := recover(); r != nil {
			sugaredLogger.Warnf("Unhandled error: %v", r)
			sugaredLogger.Warnf(string(debug.Stack()))

			if err, ok := r.(error); ok {
				resultError = err
			} else {
				panic(r)
			}
		}
	}()

	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	limit, err := e.GetFloatArgDefault(1, math.Inf(1))
	if err != nil {
		return nil, err
	}

	resultSeriesList := make([]*types.MetricData, 0, len(seriesList))
	for _, series := range seriesList {
		pointsQty := len(series.Values)
		if pointsQty != len(series.IsAbsent) {
			resultError = fmt.Errorf(
				"different Values (%d) and IsAbsent (%d) length for series %s",
				pointsQty,
				len(series.IsAbsent),
				series.Name,
			)
			return nil, resultError
		}

		resultSeries := *series
		resultSeries.Name = fmt.Sprintf("interpolate(%s)", series.Name)

		resultSeries.IsAbsent = make([]bool, pointsQty)
		copy(resultSeries.IsAbsent, series.IsAbsent)

		resultSeries.Values = make([]float64, pointsQty)
		copy(resultSeries.Values, series.Values)

		consecutiveNulls := 0
		for i := 0; i < pointsQty; i++ {
			if i == 0 {
				// no "keeping" can be done on the first value
				//because we have no idea what came before it
				continue
			}

			isAbsent := resultSeries.IsAbsent[i]
			value := resultSeries.Values[i]

			if isAbsent {
				consecutiveNulls += 1
			} else if consecutiveNulls == 0 {
				// have a value but no need to interpolate
				continue
			} else if resultSeries.IsAbsent[i-consecutiveNulls-1] {
				// # have a value but can't interpolate: reset counter
				consecutiveNulls = 0
				continue
			} else {
				// have a value and can interpolate
				// if a non-null value is seen before the limit is hit
				// backfill all the missing datapoints with the last known value
				if consecutiveNulls > 0 && float64(consecutiveNulls) <= limit {
					lastNotNullIndex := i - consecutiveNulls - 1
					lastNotNullValue := resultSeries.Values[lastNotNullIndex]

					for j := 0; j < consecutiveNulls; j++ {
						coefficient := float64(j+1) / float64(consecutiveNulls+1)
						index := i - consecutiveNulls + j

						resultSeries.IsAbsent[index] = false
						resultSeries.Values[index] = lastNotNullValue + coefficient*(value-lastNotNullValue)
					}
				}

				// reset counter
				consecutiveNulls = 0
			}
		}

		resultSeriesList = append(resultSeriesList, &resultSeries)
	}

	return resultSeriesList, nil
}

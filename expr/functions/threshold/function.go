package threshold

import (
	"fmt"
	"strings"

	"github.com/PAFomin-at-avito/zapwriter"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type threshold struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &threshold{}
	for _, n := range []string{"threshold"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *threshold) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	sugaredLogger := zapwriter.Logger("functionDo").With(zap.String("function", "threshold")).Sugar()

	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	rawThresholds, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	defaultThreshold, err := e.GetFloatArg(2)
	if err != nil {
		return nil, err
	}

	callID := uuid.NewV4().String()
	expressionString := e.ToString()
	sugaredLogger.Infow(
		"function called",
		"callID", callID,
		"expression", expressionString,
		"from", from,
		"until", until,
	)

	thresholds := make([]*types.MetricData, len(rawThresholds))
	for i, rawThreshold := range rawThresholds {
		thresholds[i] = f.prepareThreshold(rawThreshold, defaultThreshold)
	}

	var results []*types.MetricData

	for _, a := range args {
		var isTagged bool
		nameEndsAt := strings.IndexByte(a.Name, ';')
		if nameEndsAt == -1 {
			isTagged = false
		} else {
			isTagged = true
		}

		// try to find a threshold with matching tags
		var threshold *types.MetricData
		if isTagged {
			tags := a.Name[nameEndsAt:]
			for _, th := range thresholds {
				if hasTheseTags(th.Name, tags) {
					threshold = th
					break
				}
			}
		} else {
			for _, th := range thresholds {
				if th.Name == a.Name {
					threshold = th
					break
				}
			}
		}

		if threshold != nil {
			sugaredLogger.Infow(
				"found threshold for metric",
				"callID", callID,
				"metric", a.Name,
				"threshold", threshold.Name,
				"threshold.StartTime", threshold.StartTime,
				"threshold.StopTime", threshold.StopTime,
				"threshold.StepTime", threshold.StepTime,
			)
		} else {
			sugaredLogger.Infow(
				"no threshold found for metric",
				"callID", callID,
				"metric", a.Name,
			)
		}

		r := *a
		r.IsAbsent = make([]bool, len(a.Values))
		r.Values = make([]float64, len(a.Values))
		keepThisSeries := false
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.IsAbsent[i] = true
				r.Values[i] = 0
				continue
			}

			if threshold == nil {
				if v >= defaultThreshold {
					if !keepThisSeries {
						sugaredLogger.Infow(
							"metric above default threshold, return it",
							"callID", callID,
							"metric", a.Name,
							"point", v,
						)
					}

					r.Values[i] = v
					keepThisSeries = true
				} else {
					r.Values[i] = 0
				}

			} else {
				iThreshold := (int32(i) * a.StepTime) / threshold.StepTime
				if threshold.IsAbsent[iThreshold] {
					// TODO: this may need to be changed
					continue
				}

				if v >= threshold.Values[iThreshold] {
					if !keepThisSeries {
						sugaredLogger.Infow(
							"metric above custom threshold, return it",
							"callID", callID,
							"metric", a.Name,
							"point", v,
							"threshold", threshold.Name,
							"thresholdPoint", threshold.Values[iThreshold],
						)
					}

					r.Values[i] = v
					keepThisSeries = true
				} else {
					r.Values[i] = 0
				}

			}

		}

		if keepThisSeries {
			results = append(results, &r)
		}
	}

	sugaredLogger.Infow(
		"function finished",
		"callID", callID,
		"expression", expressionString,
		"from", from,
		"until", until,
	)
	return results, nil
}

// prepareThreshold acts as keepLastValue(10) | transformNull($defaultThreshold).
func (f *threshold) prepareThreshold(series *types.MetricData, defaultValue float64) *types.MetricData {
	// sugaredLogger := zapwriter.Logger("functionDo").With(zap.String("function", "threshold")).Sugar()

	r := *series
	r.Values = make([]float64, len(series.Values))
	r.IsAbsent = make([]bool, len(series.Values))

	fillValue := defaultValue
	for i := len(series.Values) - 1; i >= 0; i-- {
		if !series.IsAbsent[i] {
			fillValue = series.Values[i]
			break
		}
	}

	for i, v := range series.Values {
		if series.IsAbsent[i] {
			r.Values[i] = fillValue
		} else {
			r.Values[i] = v
		}
	}

	return &r
}

func hasTheseTags(fullMetric string, tags string) bool {
	return strings.HasSuffix(fullMetric, tags) && strings.IndexByte(strings.TrimSuffix(fullMetric, tags), ';') == -1
}

type tags map[string]string

func parseTags(metricName string) (tags, error) {
	result := make(tags)
	pieces := strings.Split(metricName, ";")
	result["name"] = pieces[0]
	for i := 1; i < len(pieces); i++ {
		keyValue := strings.SplitN(pieces[i], "=", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("malformed piece: %s", pieces[i])
		}
		result[keyValue[0]] = keyValue[1]
	}
	return result, nil
}
func (t tags) IsSubset(other tags) bool {
	for tag, value := range t {
		if otherValue, ok := other[tag]; !ok || otherValue != value {
			return false
		}
	}
	return true
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *threshold) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"threshold": {
			Description: "Compares a list of tagged metrics with a list of tagged threshold metrics.",
			Function:    "threshold(metricSeriesList, thresholdSeriesList, defaultThreshold)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "threshold",
			Params: []types.FunctionParam{
				{
					Name:     "metricSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "thresholdSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "defaultThreshold",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}

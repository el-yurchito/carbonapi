package threshold

import (
	"fmt"
	"math"
	"strings"

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
				if strings.HasSuffix(th.Name, tags) && strings.IndexByte(strings.TrimSuffix(th.Name, tags), ';') == -1 {
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

		r := *a
		r.IsAbsent = make([]bool, len(a.Values))
		r.Values = make([]float64, len(a.Values))
		keepThisSeries := false
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}

			if threshold == nil {
				if v >= defaultThreshold {
					r.Values[i] = v
					keepThisSeries = true
				} else {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}

			} else {
				iThreshold := (int32(i) * a.StepTime) / threshold.StepTime
				if threshold.IsAbsent[iThreshold] {
					// TODO: this may need to be changed
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}

				if v >= threshold.Values[iThreshold] {
					r.Values[i] = v
					keepThisSeries = true
				} else {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}

			}

		}

		if keepThisSeries {
			results = append(results, &r)
		}
	}

	return results, nil
}

// prepareThreshold acts as keepLastValue(10) | transformNull($defaultThreshold).
func (f *threshold) prepareThreshold(series *types.MetricData, defaultValue float64) *types.MetricData {
	const KeepLastValues = 10

	r := *series
	r.Values = make([]float64, len(series.Values))
	r.IsAbsent = make([]bool, len(series.Values))

	prev := math.NaN()
	missing := 0
	for i, v := range series.Values {
		if series.IsAbsent[i] {
			if (missing < KeepLastValues) && !math.IsNaN(prev) {
				r.Values[i] = prev
				missing++
			} else {
				r.Values[i] = defaultValue
			}
		} else {
			prev = v
			missing = 0
			r.Values[i] = v
		}
	}

	return &r
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

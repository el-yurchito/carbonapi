package timeShiftByMetric

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type timeShiftByMetric struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{interfaces.FunctionMetadata{
		F:    &timeShiftByMetric{},
		Name: "timeShiftByMetric",
	}}
}

func (f *timeShiftByMetric) Description() map[string]types.FunctionDescription {
	return nil
}

func (f *timeShiftByMetric) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	helper.GetSeriesArg(e.Args()[0], from, until, values)
	return nil, nil
}

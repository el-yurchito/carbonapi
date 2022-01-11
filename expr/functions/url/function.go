package url

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type urlDecode struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &urlDecode{}, Name: "urlDecode"},
	}
}

func (f *urlDecode) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"urlDecode": {
			Description: "Makes urlDecode for given nodes (zero-based index) of metric name. Decodes the whole name if indices list is empty.",
			Function:    "urlDecode(seriesList, idx1, idx2, idx3, ...)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "urlDecode",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "nodes",
					Required: false,
					Type:     types.Node,
				},
			},
		},
	}
}

func (f *urlDecode) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	indices, err := e.GetIntArgs(1)
	if err != nil && err != parser.ErrMissingArgument { // indices list can be empty
		return nil, err
	}

	for _, arg := range args {
		metric := helper.ExtractMetric(arg.Name)
		arg.Name, err = applyUrlDecode(metric, indices)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

func applyUrlDecode(metric string, indices []int) (string, error) {
	if len(indices) == 0 {
		return doUrlDecode(metric), nil
	}

	const sep = "."
	nodes := strings.Split(metric, sep)
	nodesQty := len(nodes)

	for _, idx := range indices {
		if idx < 0 {
			idx += nodesQty
		}
		if idx < 0 || idx >= nodesQty {
			return "", fmt.Errorf("bad node number %d for metric %s", idx, metric)
		}

		nodes[idx] = doUrlDecode(nodes[idx])
	}
	return strings.Join(nodes, sep), nil
}

func doUrlDecode(str string) string {
	size := len(str)
	buff := strings.Builder{}
	buff.Grow(size)

	for i := 0; i < size; i++ {
		b := str[i]
		if b == '_' {
			// `__` -> `+_`
			// `_` -> `%`
			if (i < size-1 && str[i+1] == '_') || (i == size-1) {
				b = '+'
			} else {
				b = '%'
			}
		}
		buff.WriteByte(b)
	}

	result, err := url.QueryUnescape(buff.String())
	if err != nil {
		return str
	}
	return result
}

package join

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type join struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &join{}, Name: "join"},
	}
}

func (f *join) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"join": {
			Description: `Performs set operations on 'seriesA' and 'seriesB'. Following options are available:
 * AND - returns those metrics from 'seriesA' which are presented in 'seriesB';
 * OR  - returns all metrics from 'seriesA' and also those metrics from 'seriesB' which aren't presented in 'seriesA';
 * XOR - returns only those metrics which are presented in either 'seriesA' or 'seriesB', but not in both;
 * SUB - returns those metrics from 'seriesA' which aren't presented in 'seriesB';

Example:

.. code-block:: none

  &target=join(some.data.series.aaa, some.other.series.bbb, 'AND')`,
			Function: "join(seriesA, seriesB)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "join",
			Params: []types.FunctionParam{
				{
					Name:     "seriesA",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesB",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "type",
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion(and),
					Options:  []string{and, or, xor, sub},
				},
				{
					Name:     "nodesA",
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion(""),
				},
				{
					Name:     "nodesB",
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion(""),
				},
			},
		},
	}
}

const (
	and = "AND"
	or  = "OR"
	xor = "XOR"
	sub = "SUB"
)

func (f *join) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesA, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	seriesB, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}
	joinType, err := e.GetStringNamedOrPosArgDefault("type", 2, and)
	if err != nil {
		return nil, err
	}
	joinType = strings.ToUpper(joinType)

	nodesA, err := e.GetStringNamedOrPosArgDefault("nodesA", 3, "")
	if err != nil {
		return nil, err
	}
	nodesB, err := e.GetStringNamedOrPosArgDefault("nodesB", 4, "")
	if err != nil {
		return nil, err
	}

	transformerA, err := parseTransformerArg(nodesA)
	if err != nil {
		return nil, err
	}
	transformerB, err := parseTransformerArg(nodesB)
	if err != nil {
		return nil, err
	}

	switch joinType {
	case and:
		return doAnd(seriesA, seriesB, transformerA, transformerB), nil
	case or:
		return doOr(seriesA, seriesB, transformerA, transformerB), nil
	case xor:
		return doXor(seriesA, seriesB, transformerA, transformerB), nil
	case sub:
		return doSub(seriesA, seriesB, transformerA, transformerB), nil
	default:
		return nil, fmt.Errorf("unknown join type: %s", joinType)
	}
}

func doAnd(
	seriesA, seriesB []*types.MetricData,
	transformerA, transformerB metricNameTransformer,
) (results []*types.MetricData) {
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[transformerB.transform(md.Name)] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA))
	for _, md := range seriesA {
		if metricsB[transformerA.transform(md.Name)] {
			results = append(results, md)
		}
	}
	return results
}

func doOr(
	seriesA, seriesB []*types.MetricData,
	transformerA, transformerB metricNameTransformer,
) (results []*types.MetricData) {
	metricsA := make(map[string]bool, len(seriesA))
	for _, md := range seriesA {
		metricsA[transformerA.transform(md.Name)] = true
	}

	results = seriesA
	for _, md := range seriesB {
		if !metricsA[transformerB.transform(md.Name)] {
			results = append(results, md)
		}
	}
	return results
}

func doXor(
	seriesA, seriesB []*types.MetricData,
	transformerA, transformerB metricNameTransformer,
) (results []*types.MetricData) {
	metricsA := make(map[string]bool, len(seriesA))
	for _, md := range seriesA {
		metricsA[transformerA.transform(md.Name)] = true
	}
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[transformerB.transform(md.Name)] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA)+len(seriesB))
	for _, md := range seriesA {
		if !metricsB[transformerA.transform(md.Name)] {
			results = append(results, md)
		}
	}
	for _, md := range seriesB {
		if !metricsA[transformerB.transform(md.Name)] {
			results = append(results, md)
		}
	}
	return results
}

func doSub(
	seriesA, seriesB []*types.MetricData,
	transformerA, transformerB metricNameTransformer,
) (results []*types.MetricData) {
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[transformerB.transform(md.Name)] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA))
	for _, md := range seriesA {
		if !metricsB[transformerA.transform(md.Name)] {
			results = append(results, md)
		}
	}
	return results
}

const sep = "."

type metricNameTransformer interface {
	transform(string) string
}

type noop struct{}

func (t noop) transform(name string) string {
	return name
}

type nodeNumbers []int

func (t nodeNumbers) transform(name string) string {
	result := strings.Builder{}
	parts := strings.Split(name, sep)
	partsQty := len(parts)
	nodesQty := len(t)

	for i, nodeNumber := range t {
		var part string
		if nodeNumber >= 0 && nodeNumber < partsQty {
			part = parts[nodeNumber]
		}
		if nodeNumber < 0 && nodeNumber >= -partsQty {
			part = parts[partsQty+nodeNumber]
		}

		result.WriteString(part)
		if i < nodesQty-1 {
			result.WriteString(sep)
		}
	}

	return result.String()
}

func parseNodesList(val string) (numbers nodeNumbers, err error) {
	parts := strings.Split(val, sep)
	numbers = make(nodeNumbers, 0, len(parts))

	for _, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node numbers list '%s': %w", val, err)
		}
		numbers = append(numbers, number)
	}

	return numbers, nil
}

func parseTransformerArg(val string) (transformer metricNameTransformer, err error) {
	if val == "" {
		return noop{}, nil
	}
	return parseNodesList(val)
}

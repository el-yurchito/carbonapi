package transformNull

import (
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type transformNull struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &transformNull{}
	functions := []string{"transformNull"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// Do example is: transformNull(seriesList, default=0, smoothTail=true)
func (f *transformNull) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	defv, err := e.GetFloatNamedOrPosArgDefault("default", 1, 0)
	if err != nil {
		return nil, err
	}

	smoothTail, err := e.GetBoolNamedOrPosArgDefault("smoothTail", 2, true)
	if err != nil {
		return nil, err
	}

	_, ok := e.NamedArgs()["default"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	var results []*types.MetricData
	for _, a := range arg {
		var name string
		if !smoothTail {
			name = fmt.Sprintf("transformNull(%s,%g,%v)", a.Name, defv, smoothTail)
		} else if ok {
			name = fmt.Sprintf("transformNull(%s,%g)", a.Name, defv)
		} else {
			name = fmt.Sprintf("transformNull(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values)) // all points are considered not to be absent by default

		valuesQty := len(a.Values)
		for i, val := range a.Values {
			if !a.IsAbsent[i] {
				// don't modify actual (present) points
				r.Values[i] = val
				continue
			}

			if smoothTail && i == valuesQty-1 {
				// the last point is absent
				r.IsAbsent[i] = true
			} else {
				// default value, the point isn't absent
				r.Values[i] = defv
			}
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *transformNull) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"transformNull": {
			Description: `Takes a metric or wildcard seriesList and replaces null values with the value
specified by 'default'.  The value 0 used if not specified.  
Parameter 'smoothTail' (true by default) leaves last point absent (if it is actually absent in source series).

Example:

.. code-block:: none

  &target=transformNull(webapp.pages.*.views,-1)

This would take any page that didn't have values and supply negative 1 as a default.
Any other numeric value may be used as well.`,
			Function: "transformNull(seriesList, default=0, referenceSeries=None)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "transformNull",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default:  types.NewSuggestion(0),
					Name:     "default",
					Required: false,
					Type:     types.Float,
				},
				{
					Default:  types.NewSuggestion(true),
					Name:     "smoothTail",
					Required: false,
					Type:     types.Boolean,
				},
			},
		},
	}
}

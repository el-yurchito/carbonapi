package parser

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParseExpr(t *testing.T) {

	tests := []struct {
		s string
		e *expr
	}{
		{"metric",
			&expr{target: "metric"},
		},
		{
			"metric.foo",
			&expr{target: "metric.foo"},
		},
		{"metric.*.foo",
			&expr{target: "metric.*.foo"},
		},
		{
			"func(metric)",
			&expr{
				target:     "func",
				exprType:   EtFunc,
				args:       []*expr{{target: "metric"}},
				argsString: "metric",
			},
		},
		{
			"func(metric1,metric2,metric3)",
			&expr{
				target:   "func",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "metric2"},
					{target: "metric3"}},
				argsString: "metric1,metric2,metric3",
			},
		},
		{
			"func1(metric1,func2(metricA, metricB),metric3)",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "func2",
						exprType:   EtFunc,
						args:       []*expr{{target: "metricA"}, {target: "metricB"}},
						argsString: "metricA, metricB",
					},
					{target: "metric3"}},
				argsString: "metric1,func2(metricA, metricB),metric3",
			},
		},

		{
			"3",
			&expr{val: 3, exprType: EtConst},
		},
		{
			"3.1",
			&expr{val: 3.1, exprType: EtConst},
		},
		{
			"func1(metric1, 3, 1e2, 2e-3)",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: 3, exprType: EtConst},
					{val: 100, exprType: EtConst},
					{val: 0.002, exprType: EtConst},
				},
				argsString: "metric1, 3, 1e2, 2e-3",
			},
		},
		{
			"func1(metric1, 'stringconst')",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "stringconst", exprType: EtString},
				},
				argsString: "metric1, 'stringconst'",
			},
		},
		{
			`func1(metric1, "stringconst")`,
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "stringconst", exprType: EtString},
				},
				argsString: `metric1, "stringconst"`,
			},
		},
		{
			"func1(metric1, -3)",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: -3, exprType: EtConst},
				},
				argsString: "metric1, -3",
			},
		},

		{
			"func1(metric1, -3 , 'foo' )",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: -3, exprType: EtConst},
					{valStr: "foo", exprType: EtString},
				},
				argsString: "metric1, -3 , 'foo' ",
			},
		},

		{
			`foo.{bar,baz}.qux`,
			&expr{
				target:   "foo.{bar,baz}.qux",
				exprType: EtName,
			},
		},
		{
			`foo.b[0-9].qux`,
			&expr{
				target:   "foo.b[0-9].qux",
				exprType: EtName,
			},
		},
		{
			`virt.v1.*.text-match:<foo.bar.qux>`,
			&expr{
				target:   "virt.v1.*.text-match:<foo.bar.qux>",
				exprType: EtName,
			},
		},
		{
			"func2(metricA, metricB)|func1(metric1,metric3)",
			&expr{
				target:   "func1",
				exprType: EtFunc,
				args: []*expr{
					{target: "func2",
						exprType:   EtFunc,
						args:       []*expr{{target: "metricA"}, {target: "metricB"}},
						argsString: "metricA, metricB",
					},
					{target: "metric1"},
					{target: "metric3"}},
				argsString: "func2(metricA, metricB),metric1,metric3",
			},
		},
		{
			`movingAverage(company.server*.applicationInstance.requestsHandled|aliasByNode(1),"5min")`,
			&expr{
				target:   "movingAverage",
				exprType: EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						exprType: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, exprType: EtConst},
						},
						argsString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{exprType: EtString, valStr: "5min"},
				},
				argsString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`aliasByNode(company.server*.applicationInstance.requestsHandled,1)|movingAverage("5min")`,
			&expr{
				target:   "movingAverage",
				exprType: EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						exprType: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, exprType: EtConst},
						},
						argsString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{exprType: EtString, valStr: "5min"},
				},
				argsString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`company.server*.applicationInstance.requestsHandled|aliasByNode(1)|movingAverage("5min")`,
			&expr{
				target:   "movingAverage",
				exprType: EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						exprType: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, exprType: EtConst},
						},
						argsString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{exprType: EtString, valStr: "5min"},
				},
				argsString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`company.server*.applicationInstance.requestsHandled|keepLastValue()`,
			&expr{
				target:   "keepLastValue",
				exprType: EtFunc,
				args: []*expr{
					{target: "company.server*.applicationInstance.requestsHandled"},
				},
				argsString: `company.server*.applicationInstance.requestsHandled`,
			},
		},
		{"hello&world",
			&expr{target: "hello&world"},
		},
	}

	for _, tt := range tests {
		e, _, err := ParseExpr(tt.s)
		if err != nil {
			t.Errorf("parse for %+v failed: err=%v", tt.s, err)
			continue
		}
		if !reflect.DeepEqual(e, tt.e) {
			t.Errorf("parse for %+v failed:\ngot  %+s\nwant %+v", tt.s, spew.Sdump(e), spew.Sdump(tt.e))
		}
	}
}

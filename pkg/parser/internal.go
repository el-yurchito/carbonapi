package parser

import (
	"context"
	"fmt"
	"runtime/debug"
)

func (e *expr) deepReplaceContext(ctx context.Context) {
	replace := append(make([]*expr, 0, 64), e) // expressions to replace context; first one is self
	visited := make(map[*expr]bool, 64)        // track replaced expressions to avoid duplicating

	// `replace` length may change during iteration
	// but it's ok as long as `for ... i < len(replaced) ...` is used
	for i := 0; i < len(replace); i++ {
		current := replace[i]
		current.context = ctx
		visited[current] = true

		for _, arg := range current.args {
			if !visited[arg] {
				replace = append(replace, arg)
			}
		}
		for _, arg := range current.argsNamed {
			if !visited[arg] {
				replace = append(replace, arg)
			}
		}
	}
}

func (e *expr) doGetIntArg() (int, error) {
	if e.exprType != EtConst {
		return 0, ErrBadType
	}

	return int(e.val), nil
}

func (e *expr) doGetInt32Arg() (int32, error) {
	if e.exprType != EtConst {
		return int32(0), ErrBadType
	}

	return int32(e.val), nil
}

func (e *expr) doGetInt64Arg() (int64, error) {
	if e.exprType != EtConst {
		return int64(0), ErrBadType
	}

	return int64(e.val), nil
}

func (e *expr) doGetFloatArg() (float64, error) {
	if e.exprType != EtConst {
		return 0, ErrBadType
	}

	return e.val, nil
}

func (e *expr) doGetStringArg() (string, error) {
	if e.exprType != EtString {
		return "", ErrBadType
	}

	return e.valStr, nil
}

func (e *expr) doGetBoolArg() (bool, error) {
	trg := ""
	if e.exprType == EtName {
		trg = e.target
	} else if e.exprType == EtString {
		trg = e.valStr
	} else {
		return false, ErrBadType
	}

	// names go into 'target'
	switch trg {
	case "False", "false":
		return false, nil
	case "True", "true":
		return true, nil
	}

	return false, ErrBadType
}

func (e *expr) getNamedArg(name string) *expr {
	if a, ok := e.argsNamed[name]; ok {
		return a
	}

	return nil
}

func (e *expr) toExpr() interface{} {
	return e
}

func mergeNamedArgs(arg1, arg2 map[string]*expr) map[string]*expr {
	res := make(map[string]*expr)
	if arg1 != nil {
		for k, v := range arg1 {
			res[k] = v
		}
	}
	if arg2 != nil {
		for k, v := range arg2 {
			res[k] = v
		}
	}
	return res
}

func sliceExpr(args []interface{}) ([]*expr, map[string]*expr) {
	var res []*expr
	var nArgs map[string]*expr
	for _, a := range args {
		switch v := a.(type) {
		case ArgName:
			res = append(res, NewNameExpr(string(v)).toExpr().(*expr))
		case ArgValue:
			res = append(res, NewValueExpr(string(v)).toExpr().(*expr))
		case float64:
			res = append(res, NewConstExpr(v).toExpr().(*expr))
		case int:
			res = append(res, NewConstExpr(float64(v)).toExpr().(*expr))
		case string:
			res = append(res, NewTargetExpr(v).toExpr().(*expr))
		case Expr:
			res = append(res, v.toExpr().(*expr))
		case *expr:
			res = append(res, v)
		case NamedArgs:
			nArgsNew := mapExpr(v)
			nArgs = mergeNamedArgs(nArgs, nArgsNew)
		default:
			panic(fmt.Sprintf("BUG! THIS SHOULD NEVER HAPPEN! Unknown type=%T\n%v\n", a, string(debug.Stack())))
		}
	}

	return res, nArgs
}

func mapExpr(m NamedArgs) map[string]*expr {
	if m == nil || len(m) == 0 {
		return nil
	}
	res := make(map[string]*expr)
	for k, a := range m {
		switch v := a.(type) {
		case ArgName:
			res[k] = NewNameExpr(string(v)).toExpr().(*expr)
		case ArgValue:
			res[k] = NewValueExpr(string(v)).toExpr().(*expr)
		case float64:
			res[k] = NewConstExpr(v).toExpr().(*expr)
		case int:
			res[k] = NewConstExpr(float64(v)).toExpr().(*expr)
		case string:
			res[k] = NewTargetExpr(v).toExpr().(*expr)
		case Expr:
			res[k] = v.toExpr().(*expr)
		case *expr:
			res[k] = v
		default:
			return nil
		}
	}

	return res
}

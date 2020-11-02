package expr

import (
	// Import all known functions
	_ "github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/timer"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type evaluator struct{}

// EvalExpr evalualtes expressions
func (eval evaluator) EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return EvalExpr(e, from, until, values)
}

var _evaluator = evaluator{}

func init() {
	helper.SetEvaluator(_evaluator)
	metadata.SetEvaluator(_evaluator)
}

// EvalExpr is the main expression evaluator
func EvalExpr(expr parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if expr.IsName() {
		return values[parser.MetricRequest{Metric: expr.Target(), From: from, Until: until}], nil
	} else if expr.IsConst() {
		p := types.MetricData{FetchResponse: pb.FetchResponse{Name: expr.Target(), Values: []float64{expr.FloatValue()}}}
		return []*types.MetricData{&p}, nil
	}
	// evaluate the function

	// all functions have arguments -- check we do too
	if len(expr.Args()) == 0 {
		return nil, parser.ErrMissingArgument
	}

	metadata.FunctionMD.RLock()
	f, ok := metadata.FunctionMD.Functions[expr.Target()]
	metadata.FunctionMD.RUnlock()
	if !ok {
		return nil, helper.ErrUnknownFunction(expr.Target())
	}

	// trace function call
	callStack := timer.RestoreFromContext(expr.GetContext())
	if callStack != nil {
		callStack.CallStarted(expr, from, until)
	}
	result, err := f.Do(expr, from, until, values)
	if callStack != nil {
		callStack.CallFinished(err)
	}

	return result, err
}

// RewriteExpr expands targets that use applyByNode into a new list of targets.
// eg:
// applyByNode(foo*, 1, "%") -> (true, ["foo1", "foo2"], nil)
// sumSeries(foo) -> (false, nil, nil)
// Assumes that applyByNode only appears as the outermost function.
func RewriteExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	if e.IsFunc() {
		metadata.FunctionMD.RLock()
		f, ok := metadata.FunctionMD.RewriteFunctions[e.Target()]
		metadata.FunctionMD.RUnlock()
		if ok {
			return f.Do(e, from, until, values)
		}
	}
	return false, nil, nil
}

package timer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const (
	contextKey = "FunctionCallStack"
	rootParent = int32(-1)
)

const (
	trmSecondsPerMinute = 60
	trmMinutesPerHour   = 60
	trmHoursPerDay      = 24
)

// FunctionCall contains data about single function call and its connection to the other ones
type FunctionCall = carbonapipb.FunctionCall

// FunctionCallStack contains data about all calls which where made during expression's evaluation
type FunctionCallStack struct {
	calls    []*FunctionCall
	isolated []*IsolatedCall
	current  int32
	total    int32
}

// IsolatedCall contains information about execution time of
// single function call corrected by duration of all nested calls
type IsolatedCall struct {
	From, Until   int32
	ExecutionTime int64
	Name          string
}

// TimeRangeMarks contains information about time range split to different time units
// each unit doesn't include higher units' values
// e.g. if time range is 2 hours 15 min 37 seconds then Days=0 Hours=2 Minutes=15 and Seconds=37
type TimeRangeMarks struct {
	Days    int32
	Hours   int32
	Minutes int32
	Seconds int32
	Range   int32
}

func NewFunctionCallStack() *FunctionCallStack {
	return &FunctionCallStack{
		calls:    make([]*FunctionCall, 0, 64),
		isolated: make([]*IsolatedCall, 0, 64),
		current:  rootParent,
		total:    0,
	}
}

func RestoreFromContext(ctx context.Context) *FunctionCallStack {
	if ctx == nil {
		return nil
	}

	result, _ := ctx.Value(contextKey).(*FunctionCallStack)
	return result
}

// CallStarted registers new call in stack
func (fcs *FunctionCallStack) CallStarted(expr parser.Expr, from, until int32) {
	target := expr.Target()

	fcs.calls = append(fcs.calls, &FunctionCall{
		FullExpression: expr.ToString(),
		Target:         target,
		From:           from,
		Until:          until,
		Parent:         fcs.current,
		Order:          fcs.total,
		CallStarted:    time.Now().UnixNano(),
	})
	fcs.isolated = append(fcs.isolated, &IsolatedCall{
		From:  from,
		Until: until,
		Name:  target,
	})

	fcs.current = fcs.total
	fcs.total++
}

// CallFinished marks last call as finished and defines new last call
func (fcs *FunctionCallStack) CallFinished(err error) {
	if fcs.current == rootParent {
		panic(fmt.Errorf("CallFinished on empty stack"))
	}

	// put current call completion info
	currentCall := fcs.calls[fcs.current]
	currentCall.CallFinished = time.Now().UnixNano()
	if err != nil {
		currentCall.Failed = true
		currentCall.ErrorMessage = err.Error()
	}

	// define its isolated execution time
	executionTime := currentCall.CallFinished - currentCall.CallStarted
	currentIsolated := fcs.isolated[fcs.current]
	currentIsolated.ExecutionTime += executionTime
	if currentCall.Parent != rootParent { // adjust isolated execution time for parent if there is any
		fcs.isolated[currentCall.Parent].ExecutionTime -= executionTime
	}

	// move stack top
	fcs.current = currentCall.Parent
}

// GetCalls returns aggregated calls
func (fcs *FunctionCallStack) GetCalls() []*FunctionCall {
	return fcs.calls
}

// GetIsolatedCalls calculates and returns IsolatedCall for each aggregated call
func (fcs *FunctionCallStack) GetIsolatedCalls() []*IsolatedCall {
	return fcs.isolated
}

// Store stores FunctionCallStack to the given context
func (fcs *FunctionCallStack) Store(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, contextKey, fcs)
}

func (ic *IsolatedCall) TimeRangeMarks() *TimeRangeMarks {
	delta := ic.Until - ic.From
	marks := &TimeRangeMarks{Range: delta}

	marks.Seconds = delta % trmSecondsPerMinute
	delta /= trmSecondsPerMinute

	marks.Minutes = delta % trmMinutesPerHour
	delta /= trmMinutesPerHour

	marks.Hours = delta % trmHoursPerDay
	delta /= trmHoursPerDay

	marks.Days = delta
	return marks
}

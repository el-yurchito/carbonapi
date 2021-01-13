package types

import (
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"

	"github.com/go-graphite/carbonapi/carbonzipperpb3"
)

// these 2 types are exposed in order to use their patched version
// RequestedTarget is added to FetchResponse comparing to original pb.FetchResponse
type (
	FetchResponse      = carbonzipperpb3.FetchResponseEx
	MultiFetchResponse = carbonzipperpb3.MultiFetchResponseEx
)

// CastFetchResponse converts original pb.FetchResponse to patched one
func CastFetchResponse(response pb.FetchResponse) *FetchResponse {
	return &FetchResponse{
		Name:      response.Name,
		StartTime: response.StartTime,
		StopTime:  response.StopTime,
		StepTime:  response.StepTime,
		Values:    response.Values,
		IsAbsent:  response.IsAbsent,
	}
}

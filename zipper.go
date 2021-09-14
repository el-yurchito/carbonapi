package main

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	realZipper "github.com/go-graphite/carbonzipper/zipper"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/util"
)

type zipper struct {
	z *realZipper.Zipper

	logger              *zap.Logger
	statsSender         func(*realZipper.Stats)
	ignoreClientTimeout bool
}

// The CarbonZipper interface exposes access to realZipper
// Exposes the functionality to find, get info or render metrics.
type CarbonZipper interface {
	Find(ctx context.Context, metric string) (pb.GlobResponse, error)
	Info(ctx context.Context, metric string) (map[string]pb.InfoResponse, error)
	Render(ctx context.Context, metric string, from, until int32) ([]*types.MetricData, error)
}

func newZipper(sender func(*realZipper.Stats), config *realZipper.Config, ignoreClientTimeout bool, logger *zap.Logger) *zipper {
	z := &zipper{
		z:                   realZipper.NewZipper(sender, config, logger),
		logger:              logger,
		statsSender:         sender,
		ignoreClientTimeout: ignoreClientTimeout,
	}

	return z
}

func (z zipper) Find(ctx context.Context, metric string) (pb.GlobResponse, error) {
	var pbresp pb.GlobResponse
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}

	res, stats, err := z.z.Find(newCtx, z.logger, metric)
	if err != nil {
		return pbresp, err
	}

	pbresp.Name = metric
	pbresp.Matches = res

	z.statsSender(stats)

	return pbresp, err
}

func (z zipper) Info(ctx context.Context, metric string) (map[string]pb.InfoResponse, error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}
	resp, stats, err := z.z.Info(newCtx, z.logger, metric)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %+v", err)
	}

	z.statsSender(stats)

	return resp, nil
}

func (z zipper) Render(ctx context.Context, metric string, from, until int32) ([]*types.MetricData, error) {
	var (
		newCtx = ctx
		result []*types.MetricData
	)

	if z.ignoreClientTimeout {
		newCtx = util.SetUUID(context.Background(), util.GetUUID(ctx))
	}

	resp, stats, err := z.z.Render(newCtx, z.logger, metric, from, until)
	if err != nil {
		return result, err
	}
	z.statsSender(stats)

	metricsQty := len(resp.Metrics)
	if metricsQty == 0 {
		return result, errors.New("no metrics")
	}

	for i := 0; i < metricsQty; i++ {
		result = append(result, &types.MetricData{FetchResponse: *types.CastFetchResponse(resp.Metrics[i])})
	}

	return result, nil
}

type upstreamError interface {
	error
	HttpStatus() int
}

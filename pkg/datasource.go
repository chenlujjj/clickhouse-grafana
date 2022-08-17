package main

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

func GetDatasourceServeOpts() datasource.ServeOpts {
	ds := &ClickHouseDatasource{
		im: datasource.NewInstanceManager(NewDatasourceSettings),
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

type ClickHouseDatasource struct {
	im instancemgmt.InstanceManager
}

func (ds *ClickHouseDatasource) getClient(ctx backend.PluginContext) (*ClickHouseClient, error) {
	im, err := ds.im.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &ClickHouseClient{
		settings: im.(*DatasourceSettings),
	}, nil
}

func (ds *ClickHouseDatasource) executeQuery(pluginContext backend.PluginContext, ctx context.Context, query *Query) backend.DataResponse {
	backend.Logger.Info("fuck: enter ClickHouseDatasource executeQuery")

	onErr := func(err error) backend.DataResponse {
		backend.Logger.Error(fmt.Sprintf("Datasource executeQuery error: %s", err))
		return backend.DataResponse{Error: err}
	}

	client, err := ds.getClient(pluginContext)
	if err != nil {
		return onErr(err)
	}
	sql := query.ApplyTimeRangeToQuery()
	clickhouseResponse, err := client.Query(ctx, sql)
	if err != nil {
		return onErr(err)
	}

	frames, err := clickhouseResponse.toFrames(query, client.FetchTimeZone)
	if err != nil {
		return onErr(err)
	}

	backend.Logger.Debug(fmt.Sprintf("queryResponse: %s returns %v frames", sql, len(frames)))
	return backend.DataResponse{
		Frames: frames,
	}
}

func (ds *ClickHouseDatasource) evalQuery(pluginContext backend.PluginContext, ctx context.Context, evalQuery *EvalQuery) backend.DataResponse {
	backend.Logger.Info("fuck: enter ClickHouseDatasource evalQuery")

	onErr := func(err error) backend.DataResponse {
		backend.Logger.Error(fmt.Sprintf("Datasource evalQuery error: %s", err))
		return backend.DataResponse{Error: err}
	}

	sql, err := evalQuery.ApplyMacrosAndTimeRangeToQuery()
	backend.Logger.Info("fuck: after ApplyMacrosAndTimeRangeToQuery sql: %s", sql)
	if err != nil {
		return onErr(err)
	}

	q := Query{
		From:     evalQuery.From,
		To:       evalQuery.To,
		RawQuery: sql,
	}
	return ds.executeQuery(pluginContext, ctx, &q)
}

func (ds *ClickHouseDatasource) QueryData(
	ctx context.Context,
	req *backend.QueryDataRequest,
) (*backend.QueryDataResponse, error) {
	onErr := func(err error) (*backend.QueryDataResponse, error) {
		backend.Logger.Error(fmt.Sprintf("QueryData error: %v", err))
		return nil, err
	}

	response := backend.NewQueryDataResponse()
	wg, wgCtx := errgroup.WithContext(ctx)
	for _, query := range req.Queries {
		q := Query{
			From: query.TimeRange.From,
			To:   query.TimeRange.To,
		}
		err := json.Unmarshal(query.JSON, &q)
		if err == nil {
			backend.Logger.Info("fuck: will ds.executeQuery, rawQuery: %s", q.RawQuery)
			wg.Go(func() error {
				response.Responses[q.RefId] = ds.executeQuery(req.PluginContext, wgCtx, &q)
				return nil
			})
		}
		if err != nil {
			evalQ := EvalQuery{
				From:          query.TimeRange.From,
				To:            query.TimeRange.To,
				MaxDataPoints: query.MaxDataPoints,
			}
			err := json.Unmarshal(query.JSON, &evalQ)
			if err != nil {
				return onErr(fmt.Errorf("unable to parse json %s into Query struct Error: %w", query.JSON, err))
			}
			backend.Logger.Info("fuck: will ds.evalQuery, database: %s, table: %s", evalQ.Database, evalQ.Table)
			wg.Go(func() error {
				response.Responses[evalQ.RefId] = ds.evalQuery(req.PluginContext, wgCtx, &evalQ)
				return nil
			})
		}
	}
	if err := wg.Wait(); err != nil {
		return onErr(fmt.Errorf("one of executeQuery go-routine return error: %v", err))
	}

	return response, nil
}

func (ds *ClickHouseDatasource) CheckHealth(
	ctx context.Context,
	req *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	backend.Logger.Info("fuck: ClickHouseDatasource CheckHealth")

	onErr := func(err error) (*backend.CheckHealthResult, error) {
		backend.Logger.Error(fmt.Sprintf("HealthCheck error: %v", err))
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, err
	}

	client, err := ds.getClient(req.PluginContext)
	if err != nil {
		return onErr(err)
	}
	_, err = client.Query(ctx, DefaultQuery)
	if err != nil {
		return onErr(err)
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "OK",
	}, nil
}

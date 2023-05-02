package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PAFomin-at-avito/zapwriter"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/intervalset"
	realZipper "github.com/go-graphite/carbonzipper/zipper"
	pickle "github.com/lomik/og-rek"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/date"
	"github.com/go-graphite/carbonapi/expr"
	"github.com/go-graphite/carbonapi/expr/functions/cairo/png"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/timer"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonapi/util/patternSub"
)

const (
	jsonFormat      = "json"
	treejsonFormat  = "treejson"
	pngFormat       = "png"
	csvFormat       = "csv"
	rawFormat       = "raw"
	svgFormat       = "svg"
	protobufFormat  = "protobuf"
	protobuf3Format = "protobuf3"
	pickleFormat    = "pickle"
)

func initHandlers() *http.ServeMux {
	r := http.DefaultServeMux
	r.HandleFunc("/render/", renderHandler)
	r.HandleFunc("/render", renderHandler)

	r.HandleFunc("/metrics/find/", findHandler)
	r.HandleFunc("/metrics/find", findHandler)

	r.HandleFunc("/info/", infoHandler)
	r.HandleFunc("/info", infoHandler)

	r.HandleFunc("/lb_check", lbcheckHandler)

	r.HandleFunc("/health-check", backendHealthHandler)
	r.HandleFunc("/health-check/", backendHealthHandler)

	r.HandleFunc("/version", versionHandler)
	r.HandleFunc("/version/", versionHandler)

	r.HandleFunc("/functions", functionsHandler)
	r.HandleFunc("/functions/", functionsHandler)

	r.HandleFunc("/tags", tagHandler)
	r.HandleFunc("/tags/", tagHandler)

	r.HandleFunc("/", usageHandler)
	return r
}

func writeResponse(w http.ResponseWriter, b []byte, format string, jsonp string) {

	switch format {
	case jsonFormat:
		if jsonp != "" {
			w.Header().Set("Content-Type", contentTypeJavaScript)
			w.Write([]byte(jsonp))
			w.Write([]byte{'('})
			w.Write(b)
			w.Write([]byte{')'})
		} else {
			w.Header().Set("Content-Type", contentTypeJSON)
			w.Write(b)
		}
	case protobufFormat, protobuf3Format:
		w.Header().Set("Content-Type", contentTypeProtobuf)
		w.Write(b)
	case rawFormat:
		w.Header().Set("Content-Type", contentTypeRaw)
		w.Write(b)
	case pickleFormat:
		w.Header().Set("Content-Type", contentTypePickle)
		w.Write(b)
	case csvFormat:
		w.Header().Set("Content-Type", contentTypeCSV)
		w.Write(b)
	case pngFormat:
		w.Header().Set("Content-Type", contentTypePNG)
		w.Write(b)
	case svgFormat:
		w.Header().Set("Content-Type", contentTypeSVG)
		w.Write(b)
	}
}

const (
	contentTypeJSON       = "application/json"
	contentTypeProtobuf   = "application/x-protobuf"
	contentTypeJavaScript = "text/javascript"
	contentTypeRaw        = "text/plain"
	contentTypePickle     = "application/pickle"
	contentTypePNG        = "image/png"
	contentTypeCSV        = "text/csv"
	contentTypeSVG        = "image/svg+xml"
)

type renderRequest struct {
	metric      string
	from, until int32
}

type renderResponse struct {
	data    []*types.MetricData
	error   error
	request renderRequest
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()

	uuidString := uuid.NewV4().String()
	ctx := util.SetUUID(r.Context(), uuidString)
	w.Header().Set("X-Carbonapi-UUID", uuidString)

	accessLogDetails := carbonapipb.AccessLogDetails{
		Handler:       "render",
		CarbonapiUuid: uuidString,
	}

	var (
		functionCallStacks []*timer.FunctionCallStack
		logAsError         bool
	)

	defer func() {
		deferredFunctionCallMetrics(functionCallStacks)
		deferredAccessLogging(&accessLogDetails, functionCallStacks, r, t0, logAsError)
	}()

	apiMetrics.Requests.Add(1)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")
	format := r.FormValue("format")
	template := r.FormValue("template")
	useCache := !parser.TruthyBool(r.FormValue("noCache"))

	var jsonp string

	if format == jsonFormat {
		// TODO(dgryski): check jsonp only has valid characters
		jsonp = r.FormValue("jsonp")
	}

	if format == "" && (parser.TruthyBool(r.FormValue("rawData")) || parser.TruthyBool(r.FormValue("rawdata"))) {
		format = rawFormat
	}
	if format == "" {
		format = pngFormat
	}

	cacheTimeout := config.Cache.DefaultTimeoutSec
	logger := zapwriter.Logger("render").With(
		zap.String("carbonapi_uuid", uuidString),
	)

	if tstr := r.FormValue("cacheTimeout"); tstr != "" {
		t, err := strconv.Atoi(tstr)
		if err != nil {
			logger.Error(
				"failed to parse cacheTimeout",
				zap.String("cache_string", tstr),
				zap.Error(err),
			)
		} else {
			cacheTimeout = int32(t)
		}
	}

	// make sure the cache key doesn't say noCache, because it will never hit
	r.Form.Del("noCache")

	// jsonp callback names are frequently autogenerated and hurt our cache
	r.Form.Del("jsonp")

	// Strip some cache-busters.  If you don't want to cache, use noCache=1
	r.Form.Del("_salt")
	r.Form.Del("_ts")
	r.Form.Del("_t") // Used by jquery.graphite.js

	cacheKey := r.Form.Encode()

	// normalize from and until values
	qtz := r.FormValue("tz")
	from32 := date.DateParamToEpoch(from, qtz, timeNow().Add(-24*time.Hour).Unix(), config.defaultTimeZone)
	until32 := date.DateParamToEpoch(until, qtz, timeNow().Unix(), config.defaultTimeZone)

	accessLogDetails.UseCache = useCache
	accessLogDetails.FromRaw = from
	accessLogDetails.From = from32
	accessLogDetails.UntilRaw = until
	accessLogDetails.Until = until32
	accessLogDetails.Tz = qtz
	accessLogDetails.CacheTimeout = cacheTimeout
	accessLogDetails.Format = format
	accessLogDetails.Targets = targets
	functionCallStacks = make([]*timer.FunctionCallStack, 0, len(targets))

	if useCache {
		tc := time.Now()
		response, err := config.queryCache.Get(cacheKey)
		td := time.Since(tc).Nanoseconds()
		apiMetrics.RenderCacheOverheadNS.Add(td)

		accessLogDetails.CarbonzipperResponseSizeBytes = 0
		accessLogDetails.CarbonapiResponseSizeBytes = int64(len(response))

		if err == nil {
			apiMetrics.RequestCacheHits.Add(1)
			writeResponse(w, response, format, jsonp)
			accessLogDetails.FromCache = true
			return
		}
		apiMetrics.RequestCacheMisses.Add(1)
	}

	if from32 == until32 {
		http.Error(w, "Invalid empty time range", http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "invalid empty time range"
		logAsError = true
		return
	}

	var (
		metrics []string
		results []*types.MetricData
	)

	errors := make(map[string]string)
	metricValues := make(map[parser.MetricRequest][]*types.MetricData)

	// keep information about targets being rewritten with expr.RewriteExpr
	// if there was a chain of rewrites then only the first one and the last one are kept
	// key are replaced targets, values are original ones
	targetsHistory := make(map[string]string, len(targets))

	// it's important to use for ... i < len ... here instead of for ... range ...
	// because slice's size may change during iteration
	for i := 0; i < len(targets); i++ {
		target := targets[i]
		exp, e, err := parser.ParseExpr(target)

		if err != nil || e != "" {
			msg := buildParseErrorString(target, e, err)
			http.Error(w, msg, http.StatusBadRequest)
			accessLogDetails.Reason = msg
			accessLogDetails.HttpCode = http.StatusBadRequest
			logAsError = true
			return
		}

		exp.SetCommonBoundaries(from32, until32)
		for _, m := range exp.Metrics() {
			metrics = append(metrics, m.Metric)
			mfetch := m
			mfetch.From += from32
			mfetch.Until += until32

			if _, exist := metricValues[mfetch]; exist {
				// already fetched this metric for this request
				continue
			}

			var (
				glob          pb.GlobResponse
				haveCacheData bool
			)

			if useCache {
				tc := time.Now()
				response, err := config.findCache.Get(m.Metric)
				td := time.Since(tc).Nanoseconds()
				apiMetrics.FindCacheOverheadNS.Add(td)

				if err == nil {
					err := glob.Unmarshal(response)
					haveCacheData = err == nil
				}
			}

			if haveCacheData {
				apiMetrics.FindCacheHits.Add(1)
			} else if !config.AlwaysSendGlobsAsIs {
				var err error

				apiMetrics.FindCacheMisses.Add(1)
				apiMetrics.FindRequests.Add(1)
				accessLogDetails.ZipperRequests++

				glob, err = config.zipper.Find(ctx, m.Metric)
				if err != nil {
					logger.Error(
						"find error",
						zap.String("metric", m.Metric),
						zap.Error(err),
					)
					continue
				}
				b, err := glob.Marshal()
				if err == nil {
					tc := time.Now()
					config.findCache.Set(m.Metric, b, 5*60)
					td := time.Since(tc).Nanoseconds()
					apiMetrics.FindCacheOverheadNS.Add(td)
				}
			}

			sendGlobs := config.AlwaysSendGlobsAsIs || (config.SendGlobsAsIs && len(glob.Matches) < config.MaxBatchSize)
			accessLogDetails.SendGlobs = sendGlobs

			// construct batch of requests for concurrent launch
			metricReplacements := make(map[renderRequest]patternSub.SubstituteInfo, 32)
			renderRequestBatch := make([]renderRequest, 0, 1024)

			if sendGlobs {
				for _, substituteInfo := range config.patternProcessor.ReplacePrefix(m.Metric) {
					newRenderRequest := renderRequest{
						metric: substituteInfo.MetricDst,
						from:   mfetch.From,
						until:  mfetch.Until,
					}

					metricReplacements[newRenderRequest] = substituteInfo
					renderRequestBatch = append(renderRequestBatch, newRenderRequest)
				}
			} else {
				for _, match := range glob.Matches {
					if match.IsLeaf {
						renderRequestBatch = append(renderRequestBatch, renderRequest{
							metric: match.Path,
							from:   mfetch.From,
							until:  mfetch.Until,
						})
					}
				}
			}

			renderRequestQty := int64(len(renderRequestBatch))
			renderResponseChan := make(chan renderResponse, renderRequestQty)

			accessLogDetails.ZipperRequests += renderRequestQty
			apiMetrics.RenderRequests.Add(renderRequestQty)

			// launching batch concurrently
			for _, request := range renderRequestBatch {
				config.limiter.Enter()
				go func(request renderRequest) {
					defer config.limiter.Leave()

					r, err := config.zipper.Render(ctx, request.metric, request.from, request.until)
					renderResponseChan <- renderResponse{
						data:    r,
						error:   err,
						request: request,
					}
				}(request)
			}

			// collecting batch results
			errors := make([]error, 0, renderRequestQty)
			for i := int64(0); i < renderRequestQty; i++ {
				response := <-renderResponseChan
				if response.error == nil {
					var newMetricData []*types.MetricData
					if sendGlobs {
						newMetricData = response.data
					} else {
						newMetricData = response.data[0:]
					}

					// backward replacement metric names
					if metricReplacement, exists := metricReplacements[response.request]; exists {
						for i := range newMetricData {
							newMetricData[i].Name = config.patternProcessor.RestoreMetricName(
								newMetricData[i].Name,
								metricReplacement,
							)
						}
					}

					metricValues[mfetch] = append(metricValues[mfetch], newMetricData...)
				} else {
					errors = append(errors, response.error)
				}
			}

			if len(errors) != 0 {
				logger.Error(
					"render error occurred while fetching data",
					zap.Any("errors", errors),
				)
				apiMetrics.RenderErrors.Add(int64(len(errors)))

				for i := range errors {
					if _, ok := errors[i].(realZipper.UpstreamResponse); ok {
						// NotFound status check is not required here. Zipper checks
						// such status and sets err to nil.
						apiMetrics.RenderUpstreamErrors.Add(1)
					}
				}

				// propagate upstream error, in case it has occurred
				for i := range errors {
					if err, ok := errors[i].(realZipper.UpstreamResponse); ok {
						status := util.TransformInadequateStatusCode(err.HttpStatus())
						upstream := err.Upstream()
						body := string(err.Body())
						message := fmt.Sprintf(
							"upstream %s, status %d; body:%s",
							upstream, status, body,
						)

						if format == jsonFormat {
							w.WriteHeader(status)

							type respDto struct {
								Upstream   string `json:"upstream"`
								HttpStatus int    `json:"http_status"`
								Body       string `json:"body"`
								Message    string `json:"message"`
							}
							resp := respDto{
								Upstream:   upstream,
								HttpStatus: status,
								Body:       body,
								Message:    message,
							}

							text, _ := json.Marshal(resp)
							writeResponse(w, text, format, jsonp)
						} else {
							http.Error(w, message, status)
						}

						accessLogDetails.Reason = errors[i].Error()
						accessLogDetails.HttpCode = int32(status)
						logAsError = true

						return
					}
				}
			}

			currentFetchedMetrics := metricValues[mfetch]
			expr.SortMetrics(currentFetchedMetrics, mfetch)
		}
		accessLogDetails.Metrics = metrics

		rewritten, newTargets, err := expr.RewriteExpr(exp, from32, until32, metricValues)
		if err != nil && err != parser.ErrSeriesDoesNotExist {
			errors[target] = err.Error()
			accessLogDetails.Reason = err.Error()
			logAsError = true
			return
		}

		if rewritten {
			// memorize original target for each new one
			for _, nt := range newTargets {
				if record, ok := targetsHistory[target]; ok {
					// if current target is the replacement of another target
					// then keep that another target (the original one)
					targetsHistory[nt] = record
				} else {
					// otherwise it is original target, keep the replacement
					targetsHistory[nt] = target
				}

				// append new target to common list; it will be processed as well
				targets = append(targets, nt)
			}

			// don't evaluate rewritten target, just move to the next one
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(
						"panic during eval:",
						zap.String("cache_key", cacheKey),
						zap.Any("reason", r),
						zap.Stack("stack"),
					)
					logAsError = true
				}
			}()

			callStack := timer.NewFunctionCallStack()
			functionCallStacks = append(functionCallStacks, callStack)
			exp.SetContext(callStack.Store(nil))

			expressions, err := expr.EvalExpr(exp, from32, until32, metricValues)
			if err != nil && err != parser.ErrSeriesDoesNotExist {
				errors[target] = err.Error()
				accessLogDetails.Reason = err.Error()
				logAsError = true
				return
			}

			// return requested target for each resulting metric
			// considering the fact that target might have been rewritten
			requestedTarget, ok := targetsHistory[target]
			if !ok {
				requestedTarget = target
			}
			for _, data := range expressions {
				data.RequestedTarget = requestedTarget
				results = append(results, data)
			}
		}()
	}

	var body []byte
	switch format {
	case jsonFormat:
		if maxDataPoints, _ := strconv.Atoi(r.FormValue("maxDataPoints")); maxDataPoints != 0 {
			types.ConsolidateJSON(maxDataPoints, results)
		}

		body = types.MarshalJSON(results)
	case protobufFormat, protobuf3Format:
		body, err = types.MarshalProtobuf(results, errors)
		if err != nil {
			logger.Info(
				"request failed",
				zap.Int("http_code", http.StatusInternalServerError),
				zap.String("reason", err.Error()),
				zap.Duration("runtime", time.Since(t0)),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case rawFormat:
		body = types.MarshalRaw(results)
	case csvFormat:
		body = types.MarshalCSV(results)
	case pickleFormat:
		body = types.MarshalPickle(results)
	case pngFormat:
		body = png.MarshalPNGRequest(r, results, template)
	case svgFormat:
		body = png.MarshalSVGRequest(r, results, template)
	}

	writeResponse(w, body, format, jsonp)

	if len(results) != 0 {
		tc := time.Now()
		config.queryCache.Set(cacheKey, body, cacheTimeout)
		td := time.Since(tc).Nanoseconds()
		apiMetrics.RenderCacheOverheadNS.Add(td)
	}

	accessLogDetails.HaveNonFatalErrors = len(errors) > 0
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuidString := uuid.NewV4().String()
	ctx := util.SetUUID(r.Context(), uuidString)

	format := r.FormValue("format")
	jsonp := r.FormValue("jsonp")
	query := r.FormValue("query")

	accessLogDetails := carbonapipb.AccessLogDetails{
		Handler:       "find",
		CarbonapiUuid: uuidString,
	}
	logAsError := false
	defer func() {
		deferredAccessLogging(&accessLogDetails, nil, r, t0, logAsError)
	}()

	if query == "" {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "missing parameter `query`"
		logAsError = true
		return
	}

	for k, v := range config.patternProcessor.GetDefaultSubstituteMap() {
		if strings.HasPrefix(query, k) {
			query = v + strings.TrimPrefix(query, k)
			break
		}
	}

	if format == "" {
		format = treejsonFormat
	}

	var (
		b     []byte
		err   error
		globs pb.GlobResponse
	)

	globs, err = config.zipper.Find(ctx, query)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	switch format {
	case treejsonFormat, jsonFormat:
		b, err = findTreejson(globs)
		format = jsonFormat
	case "completer":
		b, err = findCompleter(globs)
		format = jsonFormat
	case rawFormat:
		b, err = findList(globs)
		format = rawFormat
	case protobufFormat, protobuf3Format:
		b, err = globs.Marshal()
		format = protobufFormat
	case "", pickleFormat:
		var result []map[string]interface{}

		now := int32(time.Now().Unix() + 60)
		for _, metric := range globs.Matches {
			if strings.HasPrefix(metric.Path, "_tag") {
				continue
			}
			// Tell graphite-web that we have everything
			var mm map[string]interface{}
			if config.GraphiteWeb09Compatibility {
				// graphite-web 0.9.x
				mm = map[string]interface{}{
					// graphite-web 0.9.x
					"metric_path": metric.Path,
					"isLeaf":      metric.IsLeaf,
				}
			} else {
				// graphite-web 1.0
				interval := &intervalset.IntervalSet{Start: 0, End: now}
				mm = map[string]interface{}{
					"is_leaf":   metric.IsLeaf,
					"path":      metric.Path,
					"intervals": interval,
				}
			}
			result = append(result, mm)
		}

		p := bytes.NewBuffer(b)
		pEnc := pickle.NewEncoder(p)
		err = pEnc.Encode(result)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	writeResponse(w, b, format, jsonp)
}

type completer struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsLeaf string `json:"is_leaf"`
}

func findCompleter(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var complete = make([]completer, 0)

	for _, g := range globs.Matches {
		if strings.HasPrefix(g.Path, "_tag") {
			continue
		}
		c := completer{
			Path: g.Path,
		}

		if g.IsLeaf {
			c.IsLeaf = "1"
		} else {
			c.IsLeaf = "0"
		}

		i := strings.LastIndex(c.Path, ".")

		if i != -1 {
			c.Name = c.Path[i+1:]
		} else {
			c.Name = g.Path
		}

		complete = append(complete, c)
	}

	err := json.NewEncoder(&b).Encode(struct {
		Metrics []completer `json:"metrics"`
	}{
		Metrics: complete},
	)
	return b.Bytes(), err
}

func findList(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	for _, g := range globs.Matches {
		if strings.HasPrefix(g.Path, "_tag") {
			continue
		}

		var dot string
		// make sure non-leaves end in one dot
		if !g.IsLeaf && !strings.HasSuffix(g.Path, ".") {
			dot = "."
		}

		fmt.Fprintln(&b, g.Path+dot)
	}

	return b.Bytes(), nil
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuidString := uuid.NewV4().String()
	ctx := util.SetUUID(r.Context(), uuidString)

	format := r.FormValue("format")
	if format == "" {
		format = jsonFormat
	}

	accessLogDetails := carbonapipb.AccessLogDetails{
		Handler:       "info",
		CarbonapiUuid: uuidString,
		Format:        format,
	}
	logAsError := false
	defer func() {
		deferredAccessLogging(&accessLogDetails, nil, r, t0, logAsError)
	}()

	var (
		data map[string]pb.InfoResponse
		err  error
	)

	query := r.FormValue("target")
	if query == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "no target specified"
		logAsError = true
		return
	}

	if data, err = config.zipper.Info(ctx, query); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	var b []byte
	switch format {
	case jsonFormat:
		b, err = json.Marshal(data)
	case protobufFormat, protobuf3Format:
		err = fmt.Errorf("not implemented yet")
	default:
		err = fmt.Errorf("unknown format %v", format)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	w.Write(b)
	accessLogDetails.Runtime = time.Since(t0).Seconds()
	accessLogDetails.HttpCode = http.StatusOK
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	ald := carbonapipb.AccessLogDetails{Handler: "lbcheck"}
	defer deferredAccessLogging(&ald, nil, r, time.Now(), false)

	_, _ = w.Write([]byte("Ok\n"))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	ald := carbonapipb.AccessLogDetails{Handler: "version"}
	defer deferredAccessLogging(&ald, nil, r, time.Now(), false)

	if config.GraphiteWeb09Compatibility {
		_, _ = w.Write([]byte("0.9.15\n"))
	} else {
		_, _ = w.Write([]byte("1.0.0\n"))
	}
}

func backendHealthHandler(writer http.ResponseWriter, request *http.Request) {
	backend, _ := strconv.Atoi(request.FormValue("backend"))
	proxy, err := newHealthCheckProxy(backend)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	proxy.ServeHTTP(writer, request)
}

func functionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement helper for specific functions
	t0 := time.Now()
	accessLogDetails := carbonapipb.AccessLogDetails{Handler: "functions"}
	logAsError := false
	defer func() {
		deferredAccessLogging(&accessLogDetails, nil, r, t0, logAsError)
	}()

	apiMetrics.Requests.Add(1)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	grouped := false
	nativeOnly := false
	groupedStr := r.FormValue("grouped")
	prettyStr := r.FormValue("pretty")
	nativeOnlyStr := r.FormValue("nativeOnly")
	var marshaler func(interface{}) ([]byte, error)

	if groupedStr == "1" {
		grouped = true
	}

	if prettyStr == "1" {
		marshaler = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "\t")
		}
	} else {
		marshaler = json.Marshal
	}

	if nativeOnlyStr == "1" {
		nativeOnly = true
	}

	path := strings.Split(r.URL.EscapedPath(), "/")
	function := ""
	if len(path) >= 3 {
		function = path[2]
	}

	var b []byte
	if !nativeOnly {
		metadata.FunctionMD.RLock()
		if function != "" {
			b, err = marshaler(metadata.FunctionMD.Descriptions[function])
		} else if grouped {
			b, err = marshaler(metadata.FunctionMD.DescriptionsGrouped)
		} else {
			b, err = marshaler(metadata.FunctionMD.Descriptions)
		}
		metadata.FunctionMD.RUnlock()
	} else {
		metadata.FunctionMD.RLock()
		if function != "" {
			if !metadata.FunctionMD.Descriptions[function].Proxied {
				b, err = marshaler(metadata.FunctionMD.Descriptions[function])
			} else {
				err = fmt.Errorf("%v is proxied to graphite-web and nativeOnly was specified", function)
			}
		} else if grouped {
			descGrouped := make(map[string]map[string]types.FunctionDescription)
			for groupName, description := range metadata.FunctionMD.DescriptionsGrouped {
				desc := make(map[string]types.FunctionDescription)
				for f, d := range description {
					if d.Proxied {
						continue
					}
					desc[f] = d
				}
				if len(desc) > 0 {
					descGrouped[groupName] = desc
				}
			}
			b, err = marshaler(descGrouped)
		} else {
			desc := make(map[string]types.FunctionDescription)
			for f, d := range metadata.FunctionMD.Descriptions {
				if d.Proxied {
					continue
				}
				desc[f] = d
			}
			b, err = marshaler(desc)
		}
		metadata.FunctionMD.RUnlock()
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	_, _ = w.Write(b)
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	if config.tagDBProxy != nil {
		config.tagDBProxy.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	_, _ = w.Write([]byte("[]"))
}

var usageMsg = []byte(`
supported requests:
	/render/?target=
	/metrics/find/?query=
	/info/?target=
	/functions/
	/tags/
`)

func usageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(usageMsg)
}

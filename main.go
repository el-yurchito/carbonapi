package main

import (
	"bytes"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	"github.com/go-graphite/carbonzipper/cache"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/mstats"
	"github.com/go-graphite/carbonzipper/pathcache"
	realZipper "github.com/go-graphite/carbonzipper/zipper"
	"github.com/gorilla/handlers"
	"github.com/peterbourgon/g2g"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/functions/cairo/png"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/rewrite"
	"github.com/go-graphite/carbonapi/expr/timer"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/tagdb"
	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonapi/util/dnsmanager"
	"github.com/go-graphite/carbonapi/util/patternSub"

	"go.avito.ru/do/zapwriter"
)

var apiMetrics = struct {
	Requests              *expvar.Int
	RenderRequests        *expvar.Int
	RenderErrors          *expvar.Int
	RenderUpstreamErrors  *expvar.Int
	RequestCacheHits      *expvar.Int
	RequestCacheMisses    *expvar.Int
	RenderCacheOverheadNS *expvar.Int

	FindRequests        *expvar.Int
	FindCacheHits       *expvar.Int
	FindCacheMisses     *expvar.Int
	FindCacheOverheadNS *expvar.Int

	MemcacheTimeouts expvar.Func

	CacheSize  expvar.Func
	CacheItems expvar.Func
}{
	Requests: expvar.NewInt("requests"),
	// TODO: request_cache -> render_cache
	RenderRequests:        expvar.NewInt("render_requests"),
	RenderErrors:          expvar.NewInt("render_errors"),
	RenderUpstreamErrors:  expvar.NewInt("render_upstream_errors"),
	RequestCacheHits:      expvar.NewInt("request_cache_hits"),
	RequestCacheMisses:    expvar.NewInt("request_cache_misses"),
	RenderCacheOverheadNS: expvar.NewInt("render_cache_overhead_ns"),

	FindRequests: expvar.NewInt("find_requests"),

	FindCacheHits:       expvar.NewInt("find_cache_hits"),
	FindCacheMisses:     expvar.NewInt("find_cache_misses"),
	FindCacheOverheadNS: expvar.NewInt("find_cache_overhead_ns"),
}

var zipperMetrics = struct {
	FindRequests *expvar.Int
	FindErrors   *expvar.Int

	SearchRequests *expvar.Int

	RenderRequests *expvar.Int
	RenderErrors   *expvar.Int

	InfoRequests *expvar.Int
	InfoErrors   *expvar.Int

	Timeouts *expvar.Int

	CacheSize        expvar.Func
	CacheItems       expvar.Func
	SearchCacheSize  expvar.Func
	SearchCacheItems expvar.Func

	CacheMisses       *expvar.Int
	CacheHits         *expvar.Int
	SearchCacheMisses *expvar.Int
	SearchCacheHits   *expvar.Int
}{
	FindRequests: expvar.NewInt("zipper_find_requests"),
	FindErrors:   expvar.NewInt("zipper_find_errors"),

	SearchRequests: expvar.NewInt("zipper_search_requests"),

	RenderRequests: expvar.NewInt("zipper_render_requests"),
	RenderErrors:   expvar.NewInt("zipper_render_errors"),

	InfoRequests: expvar.NewInt("zipper_info_requests"),
	InfoErrors:   expvar.NewInt("zipper_info_errors"),

	Timeouts: expvar.NewInt("zipper_timeouts"),

	CacheHits:         expvar.NewInt("zipper_cache_hits"),
	CacheMisses:       expvar.NewInt("zipper_cache_misses"),
	SearchCacheHits:   expvar.NewInt("zipper_search_cache_hits"),
	SearchCacheMisses: expvar.NewInt("zipper_search_cache_misses"),
}

// BuildVersion is provided to be overridden at build time. Eg. go build -ldflags -X 'main.BuildVersion=...'
var BuildVersion = "(development build)"

var hostname string

// for testing
var timeNow = time.Now

func splitRemoteAddr(addr string) (string, string) {
	tmp := strings.Split(addr, ":")
	if len(tmp) < 1 {
		return "unknown", "unknown"
	}
	if len(tmp) == 1 {
		return tmp[0], ""
	}

	return tmp[0], tmp[1]
}

func buildParseErrorString(target, e string, err error) string {
	msg := fmt.Sprintf("%s\n\n%-20s: %s\n", http.StatusText(http.StatusBadRequest), "Target", target)
	if err != nil {
		msg += fmt.Sprintf("%-20s: %s\n", "Error", err.Error())
	}
	if e != "" {
		msg += fmt.Sprintf("%-20s: %s\n%-20s: %s\n",
			"Parsed so far", target[0:len(target)-len(e)],
			"Could not parse", e)
	}
	return msg
}

func deferredAccessLogging(
	ald *carbonapipb.AccessLogDetails,
	stacks []*timer.FunctionCallStack,
	req *http.Request,
	serverStats *realZipper.ServerResponseStat,
	reqStarted time.Time,
	logAsError bool,
) {
	if config.FunctionCalls.WriteLog && len(stacks) > 0 {
		totalCalls := make([]*timer.FunctionCall, 0, 64)
		for _, stack := range stacks {
			stackCalls := stack.GetCalls()
			if len(stackCalls) == 0 || stackCalls[0].CallFinished == 0 {
				continue
			}
			stackDuration := time.Duration(stackCalls[0].CallFinished - stackCalls[0].CallStarted)
			if stackDuration < config.FunctionCalls.Threshold {
				continue
			}

			totalCalls = append(totalCalls, stackCalls...)
		}
		ald.FunctionCalls = totalCalls
	}

	ald.Host = req.Host
	ald.PeerIp, ald.PeerPort = splitRemoteAddr(req.RemoteAddr)
	ald.Referer = req.Referer()
	ald.Runtime = time.Since(reqStarted).Seconds()
	ald.Uri = req.RequestURI
	ald.Url = req.URL.RequestURI()

	if !logAsError {
		ald.HttpCode = http.StatusOK
	}

	fieldsToLog := make([]zapcore.Field, 0, 11)
	fieldsToLog = append(fieldsToLog,
		zap.String("handler", ald.Handler),
		zap.String("carbonapi_uuid", ald.CarbonapiUuid),
		zap.String("peer_ip", ald.PeerIp),
	)
	if len(ald.Targets) > 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Strings("targets", ald.Targets),
		)
	}
	if len(ald.Metrics) > 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Strings("metrics", ald.Metrics),
		)
	}
	if ald.Runtime != 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Float64("runtime", ald.Runtime),
		)
	}
	if ald.HttpCode != 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Int32("http_code", ald.HttpCode),
		)
	}
	if ald.From != 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Int32("from", ald.From),
		)
	}
	if ald.Until != 0 {
		fieldsToLog = append(fieldsToLog,
			zap.Int32("until", ald.Until),
		)
	}
	fieldsToLog = append(fieldsToLog, zap.Any("data", *ald))

	// copy request headers to flat map
	headers := make(map[string]string, len(req.Header))
	for key := range req.Header {
		if value := req.Header.Get(key); value != "" {
			headers[key] = value
		}
	}

	// collect various source data
	sources := map[string]string{"peer_ip": ald.PeerIp}
	for _, header := range []string{
		"User-Agent", "X-Source",
		"X-Forwarded-For", "X-Real-Ip",
		"X-Dashboard-Id", "X-Panel-Id",
		"X-Bot", "X-Office", "X-Trigger-Id",
	} {
		if val, ok := headers[header]; ok {
			sources[header] = val
		}
	}
	if _, ok := sources["X-Source"]; ok {
		sources["auto_grafana"] = "false"
	} else {
		sources["auto_grafana"] = "true"
	}

	dm := dnsmanager.Get()
	sources["peer_domain_name"] = dm.GetDomainNameByIP(ald.PeerIp)
	if xRealIP, ok := sources["X-Real-Ip"]; ok {
		sources["requester_domain_name"] = dm.GetDomainNameByIP(xRealIP)
	}

	fieldsToLog = append(
		fieldsToLog,
		zap.Any("headers", headers),
		zap.Any("sources", sources),
	)
	sourcesString, err := json.Marshal(sources)
	if err == nil {
		fieldsToLog = append(fieldsToLog, zap.String("some_headers", string(sourcesString)))
	}

	if serverStats != nil {
		queryIDsString, err := json.Marshal(serverStats.QueryIDs)
		if err == nil {
			fieldsToLog = append(fieldsToLog, zap.String("query_ids", string(queryIDsString)))
		}

		serverStatsStr, err := json.Marshal(serverStats.Stat)
		if err == nil {
			fieldsToLog = append(fieldsToLog, zap.String("server_stat", string(serverStatsStr)))
		}
	}

	logger := zapwriter.Logger("access")
	if logAsError {
		logger.Error("request failed", fieldsToLog...)
	} else {
		logger.Info("request served", fieldsToLog...)
	}
}

func deferredFunctionCallMetrics(functionCallStacks []*timer.FunctionCallStack) {
	if functionCallStacks == nil || !config.FunctionCalls.SendMetrics || statsdLimiter == nil {
		return
	}

	client := statsdLimiter.Get()
	defer client.Release()

	for _, stack := range functionCallStacks {
		for _, call := range stack.GetIsolatedCalls() {
			if time.Duration(call.ExecutionTime) < config.FunctionCalls.Threshold {
				continue
			}

			timeRangeMarks := call.TimeRangeMarks()
			tags := fmt.Sprintf(
				";hostname=%s;function=%s;range.requested=%d;range.days=%d;range.hours=%d;range.minutes=%d;range.seconds=%d",
				hostname,
				call.Name,
				timeRangeMarks.Range,
				timeRangeMarks.Days,
				timeRangeMarks.Hours,
				timeRangeMarks.Minutes,
				timeRangeMarks.Seconds,
			)

			client.Timing("function-calls.time.duration"+tags, call.ExecutionTime)
			client.Timing("function-calls.time.interval"+tags, timeRangeMarks.Range)
		}
	}
}

func deferredRenderMetrics(startTime time.Time) {
	client := statsdLimiter.Get()
	defer client.Release()

	tags := fmt.Sprintf(
		";hostname=%s",
		hostname,
	)

	client.Timing("render.timing_ms"+tags, time.Since(startTime).Milliseconds())
}

type treejson struct {
	AllowChildren int            `json:"allowChildren"`
	Expandable    int            `json:"expandable"`
	Leaf          int            `json:"leaf"`
	ID            string         `json:"id"`
	Text          string         `json:"text"`
	Context       map[string]int `json:"context"` // unused
}

var treejsonContext = make(map[string]int)

func findTreejson(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var tree = make([]treejson, 0)

	seen := make(map[string]struct{})

	basepath := globs.Name

	if i := strings.LastIndex(basepath, "."); i != -1 {
		basepath = basepath[:i+1]
	} else {
		basepath = ""
	}

	for _, g := range globs.Matches {
		if strings.HasPrefix(g.Path, "_tag") {
			continue
		}
		name := g.Path

		if i := strings.LastIndex(name, "."); i != -1 {
			name = name[i+1:]
		}

		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		t := treejson{
			ID:      basepath + name,
			Context: treejsonContext,
			Text:    name,
		}

		if g.IsLeaf {
			t.Leaf = 1
		} else {
			t.AllowChildren = 1
			t.Expandable = 1
		}

		tree = append(tree, t)
	}

	err := json.NewEncoder(&b).Encode(tree)
	return b.Bytes(), err
}

var defaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

type cacheConfig struct {
	Type              string   `mapstructure:"type"`
	Size              int      `mapstructure:"size_mb"`
	MemcachedServers  []string `mapstructure:"memcachedServers"`
	DefaultTimeoutSec int32    `mapstructure:"defaultTimeoutSec"`
}

type functionCallsConfig struct {
	SendMetrics bool
	WriteLog    bool
	Threshold   time.Duration
}

type graphiteConfig struct {
	Pattern  string
	Host     string
	Interval time.Duration
	Prefix   string
}

type rewriteConfig struct {
	From string
	To   string
}

type statsdConfig struct {
	Enabled bool
	Address string
	Prefix  string
}

type upstreamsConfig struct {
	realZipper.Config `mapstructure:",squash"`
	HealthChecks      []string `mapstructure:"health_checks"`
}

var config = struct {
	ExtrapolateExperiment      bool                `mapstructure:"extrapolateExperiment"`
	Logger                     []zapwriter.Config  `mapstructure:"logger"`
	Listen                     string              `mapstructure:"listen"`
	Concurency                 int                 `mapstructure:"concurency"`
	Cache                      cacheConfig         `mapstructure:"cache"`
	Cpus                       int                 `mapstructure:"cpus"`
	TimezoneString             string              `mapstructure:"tz"`
	UnicodeRangeTables         []string            `mapstructure:"unicodeRangeTables"`
	Graphite                   graphiteConfig      `mapstructure:"graphite"`
	Statsd                     statsdConfig        `mapstructure:"statsd"`
	IdleConnections            int                 `mapstructure:"idleConnections"`
	PidFile                    string              `mapstructure:"pidFile"`
	SendGlobsAsIs              bool                `mapstructure:"sendGlobsAsIs"`
	AlwaysSendGlobsAsIs        bool                `mapstructure:"alwaysSendGlobsAsIs"`
	MaxBatchSize               int                 `mapstructure:"maxBatchSize"`
	Zipper                     string              `mapstructure:"zipper"`
	Upstreams                  upstreamsConfig     `mapstructure:"upstreams"`
	ExpireDelaySec             int32               `mapstructure:"expireDelaySec"`
	GraphiteWeb09Compatibility bool                `mapstructure:"graphite09compat"`
	IgnoreClientTimeout        bool                `mapstructure:"ignoreClientTimeout"`
	FunctionCalls              functionCallsConfig `mapstructure:"functionCalls"`
	DefaultColors              map[string]string   `mapstructure:"defaultColors"`
	GraphTemplates             string              `mapstructure:"graphTemplates"`
	FunctionsConfigs           map[string]string   `mapstructure:"functionsConfig"`
	Rewrite                    []rewriteConfig     `mapstructure:"rewrite"`
	TagDB                      tagdb.Config        `mapstructure:"tagDB"`

	queryCache cache.BytesCache
	findCache  cache.BytesCache

	defaultTimeZone *time.Location

	// Zipper is API entry to carbonzipper
	zipper CarbonZipper

	// patternProcessor rewrites queries before sending them to backend
	patternProcessor *patternSub.PatternProcessor

	// Limiter limits concurrent zipper requests
	limiter util.SimpleLimiter

	tagDBProxy *tagdb.Http
}{
	ExtrapolateExperiment: false,
	Listen:                "[::]:8081",
	Concurency:            20,
	SendGlobsAsIs:         false,
	AlwaysSendGlobsAsIs:   false,
	MaxBatchSize:          100,
	Cache: cacheConfig{
		Type:              "mem",
		DefaultTimeoutSec: 60,
	},
	TimezoneString: "",
	Graphite: graphiteConfig{
		Pattern:  "{prefix}.{fqdn}",
		Host:     "",
		Interval: 60 * time.Second,
		Prefix:   "carbon.api",
	},
	Cpus:            0,
	IdleConnections: 10,
	PidFile:         "",

	queryCache: cache.NullCache{},
	findCache:  cache.NullCache{},

	defaultTimeZone: time.Local,
	Logger:          []zapwriter.Config{defaultLoggerConfig},

	Upstreams: upstreamsConfig{
		Config: realZipper.Config{
			Timeouts: realZipper.Timeouts{
				Global:       10000 * time.Second,
				AfterStarted: 2 * time.Second,
				Connect:      200 * time.Millisecond,
			},
			KeepAliveInterval: 30 * time.Second,

			MaxIdleConnsPerHost: 100,
		},
	},
	ExpireDelaySec:             10 * 60,
	GraphiteWeb09Compatibility: false,
	FunctionCalls: functionCallsConfig{
		SendMetrics: true,
		WriteLog:    false,
		Threshold:   100 * time.Millisecond,
	},

	TagDB: tagdb.Config{
		MaxConcurrentConnections: 10,
		MaxTries:                 3,
		Timeout:                  60 * time.Second,
		KeepAliveInterval:        30 * time.Second,
	},
}

func zipperStats(stats *realZipper.Stats) {
	zipperMetrics.Timeouts.Add(stats.Timeouts)
	zipperMetrics.FindErrors.Add(stats.FindErrors)
	zipperMetrics.RenderErrors.Add(stats.RenderErrors)
	zipperMetrics.InfoErrors.Add(stats.InfoErrors)
	zipperMetrics.SearchRequests.Add(stats.SearchRequests)
	zipperMetrics.SearchCacheHits.Add(stats.SearchCacheHits)
	zipperMetrics.SearchCacheMisses.Add(stats.SearchCacheMisses)
	zipperMetrics.CacheMisses.Add(stats.CacheMisses)
	zipperMetrics.CacheHits.Add(stats.CacheHits)
}

var graphTemplates map[string]png.PictureParams

func setUpConfig(logger *zap.Logger, zipper CarbonZipper) {
	config.Cache.MemcachedServers = viper.GetStringSlice("cache.memcachedServers")
	if n := viper.GetString("logger.logger"); n != "" {
		config.Logger[0].Logger = n
	}
	if n := viper.GetString("logger.file"); n != "" {
		config.Logger[0].File = n
	}
	if n := viper.GetString("logger.level"); n != "" {
		config.Logger[0].Level = n
	}
	if n := viper.GetString("logger.encoding"); n != "" {
		config.Logger[0].Encoding = n
	}
	if n := viper.GetString("logger.encodingtime"); n != "" {
		config.Logger[0].EncodingTime = n
	}
	if n := viper.GetString("logger.encodingduration"); n != "" {
		config.Logger[0].EncodingDuration = n
	}
	err := zapwriter.ApplyConfig(config.Logger)
	if err != nil {
		logger.Fatal("failed to initialize logger with requested configuration",
			zap.Any("configuration", config.Logger),
			zap.Error(err),
		)
	}

	if config.GraphTemplates != "" {
		graphTemplates = make(map[string]png.PictureParams)
		graphTemplatesViper := viper.New()
		b, err := ioutil.ReadFile(config.GraphTemplates)
		if err != nil {
			logger.Fatal("error reading graphTemplates file",
				zap.String("graphTemplate_path", config.GraphTemplates),
				zap.Error(err),
			)
		}

		if strings.HasSuffix(config.GraphTemplates, ".toml") {
			logger.Info("will parse config as toml",
				zap.String("graphTemplate_path", config.GraphTemplates),
			)
			graphTemplatesViper.SetConfigType("TOML")
		} else {
			logger.Info("will parse config as yaml",
				zap.String("graphTemplate_path", config.GraphTemplates),
			)
			graphTemplatesViper.SetConfigType("YAML")
		}

		err = graphTemplatesViper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.String("graphTemplate_path", config.GraphTemplates),
				zap.Error(err),
			)
		}

		for k := range graphTemplatesViper.AllSettings() {
			// we need to explicitly copy	YDivisors and ColorList
			newStruct := png.DefaultParams
			newStruct.ColorList = nil
			newStruct.YDivisors = nil
			sub := graphTemplatesViper.Sub(k)
			sub.Unmarshal(&newStruct)
			if newStruct.ColorList == nil || len(newStruct.ColorList) == 0 {
				newStruct.ColorList = make([]string, len(png.DefaultParams.ColorList))
				for i, v := range png.DefaultParams.ColorList {
					newStruct.ColorList[i] = v
				}
			}
			if newStruct.YDivisors == nil || len(newStruct.YDivisors) == 0 {
				newStruct.YDivisors = make([]float64, len(png.DefaultParams.YDivisors))
				for i, v := range png.DefaultParams.YDivisors {
					newStruct.YDivisors[i] = v
				}
			}
			graphTemplates[k] = newStruct
		}

		for name, params := range graphTemplates {
			png.SetTemplate(name, params)
		}
	}

	if config.DefaultColors != nil {
		for name, color := range config.DefaultColors {
			err = png.SetColor(name, color)
			if err != nil {
				logger.Warn("invalid color specified and will be ignored",
					zap.String("reason", "color must be valid hex rgb or rbga value, e.x. '#c80032', 'c80032', 'c80032ff', etc."),
					zap.Error(err),
				)
			}
		}
	}

	if config.FunctionsConfigs != nil {
		logger.Info("extra configuration for functions found",
			zap.Any("extra_config", config.FunctionsConfigs),
		)
	} else {
		config.FunctionsConfigs = make(map[string]string)
	}

	rewrite.New(config.FunctionsConfigs)
	functions.New(config.FunctionsConfigs)

	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewString("BuildVersion").Set(BuildVersion)
	expvar.Publish("config", expvar.Func(func() interface{} { return config }))

	config.limiter = util.NewSimpleLimiter(config.Concurency)

	configRewriteMap := make(map[string]string, len(config.Rewrite))
	for _, rewriteEntry := range config.Rewrite {
		configRewriteMap[rewriteEntry.From] = rewriteEntry.To
	}
	config.patternProcessor = patternSub.NewPatternProcessor(configRewriteMap)

	if config.TagDB.Url != "" {
		config.tagDBProxy, err = tagdb.NewHttp(&config.TagDB, config.patternProcessor)
		if err != nil {
			logger.Warn("failed to initialize http tag db",
				zap.String("reason", "invalid url"),
				zap.Error(err),
			)
		}
	}

	config.zipper = zipper

	switch config.Cache.Type {
	case "memcache":
		if len(config.Cache.MemcachedServers) == 0 {
			logger.Fatal("memcache cache requested but no memcache servers provided")
		}

		logger.Info("memcached configured",
			zap.Strings("servers", config.Cache.MemcachedServers),
		)
		config.queryCache = cache.NewMemcached("capi", config.Cache.MemcachedServers...)
		// find cache is only used if SendGlobsAsIs is false.
		if !config.SendGlobsAsIs {
			config.findCache = cache.NewExpireCache(0)
		}

		mcache := config.queryCache.(*cache.MemcachedCache)

		apiMetrics.MemcacheTimeouts = expvar.Func(func() interface{} {
			return mcache.Timeouts()
		})
		expvar.Publish("memcache_timeouts", apiMetrics.MemcacheTimeouts)

	case "mem":
		config.queryCache = cache.NewExpireCache(uint64(config.Cache.Size * 1024 * 1024))

		// find cache is only used if SendGlobsAsIs is false.
		if !config.SendGlobsAsIs {
			config.findCache = cache.NewExpireCache(0)
		}

		qcache := config.queryCache.(*cache.ExpireCache)

		apiMetrics.CacheSize = expvar.Func(func() interface{} {
			return qcache.Size()
		})
		expvar.Publish("cache_size", apiMetrics.CacheSize)

		apiMetrics.CacheItems = expvar.Func(func() interface{} {
			return qcache.Items()
		})
		expvar.Publish("cache_items", apiMetrics.CacheItems)

	case "null":
		// defaults
		config.queryCache = cache.NullCache{}
		config.findCache = cache.NullCache{}
	default:
		logger.Error("unknown cache type",
			zap.String("cache_type", config.Cache.Type),
			zap.Strings("known_cache_types", []string{"null", "mem", "memcache"}),
		)
	}

	if config.TimezoneString != "" {
		fields := strings.Split(config.TimezoneString, ",")

		if len(fields) != 2 {
			logger.Fatal("unexpected amount of fields in tz",
				zap.String("timezone_string", config.TimezoneString),
				zap.Int("fields_got", len(fields)),
				zap.Int("fields_expected", 2),
			)
		}

		offs, err := strconv.Atoi(fields[1])
		if err != nil {
			logger.Fatal("unable to parse seconds",
				zap.String("field[1]", fields[1]),
				zap.Error(err),
			)
		}

		config.defaultTimeZone = time.FixedZone(fields[0], offs)
		logger.Info("using fixed timezone",
			zap.String("timezone", config.defaultTimeZone.String()),
			zap.Int("offset", offs),
		)
	}

	if len(config.UnicodeRangeTables) != 0 {
		for _, stringRange := range config.UnicodeRangeTables {
			parser.RangeTables = append(parser.RangeTables, unicode.Scripts[stringRange])
		}
	} else {
		parser.RangeTables = append(parser.RangeTables, unicode.Latin)
	}

	if config.Cpus != 0 {
		runtime.GOMAXPROCS(config.Cpus)
	}

	var host string
	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || config.Graphite.Host != "" {
		switch {
		case envhost != ":" && config.Graphite.Host != "":
			host = config.Graphite.Host
		case envhost != ":":
			host = envhost
		case config.Graphite.Host != "":
			host = config.Graphite.Host
		}
	}

	hostname, _ = os.Hostname()
	hostname = strings.Replace(hostname, ".", "_", -1)

	logger.Info("starting carbonapi",
		zap.String("build_version", BuildVersion),
		zap.Any("config", config),
	)

	if host != "" {
		prefix := config.Graphite.Prefix
		pattern := config.Graphite.Pattern
		pattern = strings.Replace(pattern, "{prefix}", prefix, -1)
		pattern = strings.Replace(pattern, "{fqdn}", hostname, -1)

		// register our metrics with graphite
		graphite := g2g.NewGraphite(host, config.Graphite.Interval, 10*time.Second)

		graphite.Register(fmt.Sprintf("%s.requests", pattern), apiMetrics.Requests)
		graphite.Register(fmt.Sprintf("%s.request_cache_hits", pattern), apiMetrics.RequestCacheHits)
		graphite.Register(fmt.Sprintf("%s.request_cache_misses", pattern), apiMetrics.RequestCacheMisses)
		graphite.Register(fmt.Sprintf("%s.request_cache_overhead_ns", pattern), apiMetrics.RenderCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.find_requests", pattern), apiMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.find_cache_hits", pattern), apiMetrics.FindCacheHits)
		graphite.Register(fmt.Sprintf("%s.find_cache_misses", pattern), apiMetrics.FindCacheMisses)
		graphite.Register(fmt.Sprintf("%s.find_cache_overhead_ns", pattern), apiMetrics.FindCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.render_requests", pattern), apiMetrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.render_errors", pattern), apiMetrics.RenderErrors)
		graphite.Register(fmt.Sprintf("%s.render_upstream_errors", pattern), apiMetrics.RenderUpstreamErrors)

		if apiMetrics.MemcacheTimeouts != nil {
			graphite.Register(fmt.Sprintf("%s.memcache_timeouts", pattern), apiMetrics.MemcacheTimeouts)
		}

		if apiMetrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("%s.cache_size", pattern), apiMetrics.CacheSize)
			graphite.Register(fmt.Sprintf("%s.cache_items", pattern), apiMetrics.CacheItems)
		}

		graphite.Register(fmt.Sprintf("%s.zipper.find_requests", pattern), zipperMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.find_errors", pattern), zipperMetrics.FindErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.render_requests", pattern), zipperMetrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.render_errors", pattern), zipperMetrics.RenderErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.info_requests", pattern), zipperMetrics.InfoRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.info_errors", pattern), zipperMetrics.InfoErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.timeouts", pattern), zipperMetrics.Timeouts)

		graphite.Register(fmt.Sprintf("%s.zipper.cache_size", pattern), zipperMetrics.CacheSize)
		graphite.Register(fmt.Sprintf("%s.zipper.cache_items", pattern), zipperMetrics.CacheItems)

		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_size", pattern), zipperMetrics.SearchCacheSize)
		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_items", pattern), zipperMetrics.SearchCacheItems)

		graphite.Register(fmt.Sprintf("%s.zipper.cache_hits", pattern), zipperMetrics.CacheHits)
		graphite.Register(fmt.Sprintf("%s.zipper.cache_misses", pattern), zipperMetrics.CacheMisses)

		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_hits", pattern), zipperMetrics.SearchCacheHits)
		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_misses", pattern), zipperMetrics.SearchCacheMisses)

		graphite.Register(fmt.Sprintf("%s.dns.cache_misses", pattern), dnsmanager.DNSMetrics.CacheMisses)
		graphite.Register(fmt.Sprintf("%s.dns.lookup_addr_attempts", pattern), dnsmanager.DNSMetrics.LookupAddrAttempts)
		graphite.Register(fmt.Sprintf("%s.dns.lookup_addr_success", pattern), dnsmanager.DNSMetrics.LookupAddrSuccess)
		graphite.Register(fmt.Sprintf("%s.dns.lookup_addr_errors", pattern), dnsmanager.DNSMetrics.LookupAddrErrors)

		go mstats.Start(config.Graphite.Interval)

		graphite.Register(fmt.Sprintf("%s.alloc", pattern), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.total_alloc", pattern), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.num_gc", pattern), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.pause_ns", pattern), &mstats.PauseNS)

	}

	statsdLimiter, err = NewStatsdLimiter(config.Statsd)
	if err != nil {
		panic(fmt.Errorf("failed to initialize statsd: %w", err))
	}

	if config.PidFile != "" {
		pidfile.SetPidfilePath(config.PidFile)
		err := pidfile.Write()
		if err != nil {
			logger.Fatal(
				"error during pidfile.Write()",
				zap.Error(err),
			)
		}
	}

	helper.ExtrapolatePoints = config.ExtrapolateExperiment
	if config.ExtrapolateExperiment {
		logger.Warn("extraploation experiment is enabled",
			zap.String("reason", "this feature is highly experimental and untested"),
		)
	}
}

func setUpViper(logger *zap.Logger, configPath *string, viperPrefix string) {
	if *configPath != "" {
		b, err := ioutil.ReadFile(*configPath)
		if err != nil {
			logger.Fatal("error reading config file",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}

		if strings.HasSuffix(*configPath, ".toml") {
			logger.Info("will parse config as toml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("TOML")
		} else {
			logger.Info("will parse config as yaml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("YAML")
		}
		err = viper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}
	}

	if viperPrefix != "" {
		viper.SetEnvPrefix(viperPrefix)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("tz", "carbonapi_tz")
	viper.SetDefault("listen", "localhost:8081")
	viper.SetDefault("concurency", 20)
	viper.SetDefault("cache.type", "mem")
	viper.SetDefault("cache.size_mb", 0)
	viper.SetDefault("cache.defaultTimeoutSec", 60)
	viper.SetDefault("cache.memcachedServers", []string{})
	viper.SetDefault("cpus", 0)
	viper.SetDefault("tz", "")
	viper.SetDefault("sendGlobsAsIs", false)
	viper.SetDefault("AlwaysSendGlobsAsIs", false)
	viper.SetDefault("maxBatchSize", 100)
	viper.SetDefault("graphite.host", "")
	viper.SetDefault("graphite.interval", "60s")
	viper.SetDefault("graphite.prefix", "carbon.api")
	viper.SetDefault("graphite.pattern", "{prefix}.{fqdn}")
	viper.SetDefault("idleConnections", 10)
	viper.SetDefault("pidFile", "")
	viper.SetDefault("upstreams.buckets", 10)
	viper.SetDefault("upstreams.timeouts.global", "10s")
	viper.SetDefault("upstreams.timeouts.afterStarted", "2s")
	viper.SetDefault("upstreams.timeouts.connect", "200ms")
	viper.SetDefault("upstreams.concurrencyLimit", 0)
	viper.SetDefault("upstreams.keepAliveInterval", "30s")
	viper.SetDefault("upstreams.maxIdleConnsPerHost", 100)
	viper.SetDefault("upstreams.backends", []string{"http://127.0.0.1:8080"})
	viper.SetDefault("upstreams.carbonsearch.backend", "")
	viper.SetDefault("upstreams.carbonsearch.prefix", "virt.v1.*")
	viper.SetDefault("upstreams.graphite09compat", false)
	viper.SetDefault("expireDelaySec", 10)
	viper.SetDefault("logger", map[string]string{})
	viper.AutomaticEnv()

	err := viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.Error(err),
		)
	}
}

func setUpConfigUpstreams(logger *zap.Logger) {
	if config.Zipper != "" {
		logger.Warn("found legacy 'zipper' option, will use it instead of any 'upstreams' specified. This will be removed in future versions!")

		config.Upstreams.Backends = []string{config.Zipper}
		config.Upstreams.ConcurrencyLimitPerServer = config.Concurency
		config.Upstreams.MaxIdleConnsPerHost = config.IdleConnections
		config.Upstreams.KeepAliveInterval = 10 * time.Second
		// To emulate previous behavior
		config.Upstreams.Timeouts = realZipper.Timeouts{
			Connect:      1 * time.Second,
			AfterStarted: 600 * time.Second,
			Global:       600 * time.Second,
		}
	}
	if len(config.Upstreams.Backends) == 0 {
		logger.Fatal("no backends specified for upstreams!")
	}

	// Setup in-memory path cache for carbonzipper requests
	config.Upstreams.PathCache = pathcache.NewPathCache(config.ExpireDelaySec)
	config.Upstreams.SearchCache = pathcache.NewPathCache(config.ExpireDelaySec)

	zipperMetrics.CacheSize = expvar.Func(func() interface{} { return config.Upstreams.PathCache.ECSize() })
	expvar.Publish("cacheSize", zipperMetrics.CacheSize)

	zipperMetrics.CacheItems = expvar.Func(func() interface{} { return config.Upstreams.PathCache.ECItems() })
	expvar.Publish("cacheItems", zipperMetrics.CacheItems)

	zipperMetrics.SearchCacheSize = expvar.Func(func() interface{} { return config.Upstreams.SearchCache.ECSize() })
	expvar.Publish("searchCacheSize", zipperMetrics.SearchCacheSize)

	zipperMetrics.SearchCacheItems = expvar.Func(func() interface{} { return config.Upstreams.SearchCache.ECItems() })
	expvar.Publish("searchCacheItems", zipperMetrics.SearchCacheItems)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := zapwriter.ApplyConfig([]zapwriter.Config{defaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")
	}
	logger := zapwriter.Logger("main")

	configPath := flag.String("config", "", "Path to the `config file`.")
	envPrefix := flag.String("envprefix", "CARBONAPI_", "Preifx for environment variables override")
	if *envPrefix == "" {
		logger.Warn("empty prefix is not recommended due to possible collisions with OS environment variables")
	}
	flag.Parse()
	setUpViper(logger, configPath, *envPrefix)
	setUpConfigUpstreams(logger)
	zipper := newZipper(zipperStats, &config.Upstreams.Config, config.IgnoreClientTimeout, logger.With(zap.String("handler", "zipper")))
	setUpConfig(logger, zipper)

	r := initHandlers()
	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	err = gracehttp.Serve(&http.Server{
		Addr:    config.Listen,
		Handler: handler,
	})

	if err != nil {
		logger.Fatal("gracehttp failed",
			zap.Error(err),
		)
	}
}

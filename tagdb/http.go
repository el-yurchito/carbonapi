package tagdb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/PAFomin-at-avito/zapwriter"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonapi/util/patternSub"
)

type Http struct {
	logger           *zap.Logger
	proxy            *httputil.ReverseProxy
	limiter          util.SimpleLimiter
	patternProcessor *patternSub.PatternProcessor
}

type Config struct {
	MaxConcurrentConnections int
	MaxTries                 int
	Timeout                  time.Duration
	KeepAliveInterval        time.Duration
	Url                      string
	User                     string
	Password                 string
	ForwardHeaders           bool
}

func modifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		// TODO log msg
		resp.StatusCode = 200
		resp.Body = ioutil.NopCloser(bytes.NewBufferString(""))
		resp.ContentLength = 0
	}
	return nil
}

func NewHttp(cfg *Config, patternProcessor *patternSub.PatternProcessor) (*Http, error) {
	if cfg.Url == "" {
		// TODO log msg
		return nil, nil
	}
	target, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	result := &Http{
		logger:           zapwriter.Logger("tags"),
		limiter:          util.NewSimpleLimiter(cfg.MaxConcurrentConnections),
		patternProcessor: patternProcessor,
		proxy:            proxy,
	}

	proxy.Transport = &http.Transport{
		MaxIdleConnsPerHost: cfg.MaxConcurrentConnections,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: cfg.KeepAliveInterval,
			DualStack: true,
		}).DialContext,
	}
	proxy.ModifyResponse = modifyResponse

	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		if !cfg.ForwardHeaders {
			req.Header = http.Header{}
		}
		if req.Header.Get("Authorization") == "" && cfg.User != "" && cfg.Password != "" {
			req.SetBasicAuth(cfg.User, cfg.Password)
		}
		result.modifyRequest(req)
	}

	return result, nil
}

func (h *Http) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.limiter.Enter()
	h.proxy.ServeHTTP(w, r)
	h.limiter.Leave()
}

func (h *Http) modifyRequest(req *http.Request) {
	form, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		h.logger.Error(
			"Failed to parse form",
			zap.String("rawQuery", req.URL.RawQuery),
			zap.Error(err),
		)
		return
	}

	expr, ok := form["expr"]
	if !ok {
		return
	}

	changed := false
	modifiedExpr := make([]string, 0, len(expr))

	for _, term := range expr {
		name, value, sign, err := patternSub.SplitArgTerm(term)
		if err != nil {
			h.logger.Error(
				"Failed to parse expr term",
				zap.String("term", term),
				zap.Error(err),
			)
			return
		}

		replaces := h.patternProcessor.ReplacePrefix(value)
		if len(replaces) != 1 {
			h.logger.Error(fmt.Sprintf("Got %d replacements for term \"%s\" (must be 1)", len(replaces), term))
			return
		}

		replace := replaces[0]
		if replace.IsReplaced {
			changed = true
		}

		modifiedExpr = append(modifiedExpr, fmt.Sprintf("%s%s%s", name, sign, replace.MetricDst))
	}

	if changed {
		form["expr"] = modifiedExpr
		req.URL.RawQuery = form.Encode()
	}
}

package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

var (
	hcMu       sync.RWMutex
	hcHandlers map[int]http.Handler
)

func init() {
	hcHandlers = make(map[int]http.Handler)
}

// newHealthCheckProxy gets a proxy for given backend number
// proxy just forwards health check request to the specified backend
func newHealthCheckProxy(backend int) (http.Handler, error) {
	if backend < 0 || backend >= len(config.Upstreams.HealthChecks) {
		return nil, fmt.Errorf("invalid backend number %d", backend)
	}

	checkUrl := config.Upstreams.HealthChecks[backend]
	if checkUrl == "" {
		return nil, fmt.Errorf("health check url for backend number %d is not set", backend)
	}

	hcMu.RLock()
	handler, ok := hcHandlers[backend]
	hcMu.RUnlock()
	if ok {
		return handler, nil
	}

	hcMu.Lock()
	defer hcMu.Unlock()

	parsed, err := url.Parse(checkUrl)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)
	proxy.Transport = &http.Transport{
		MaxIdleConnsPerHost: 16,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	hcHandlers[backend] = proxy
	return proxy, nil
}

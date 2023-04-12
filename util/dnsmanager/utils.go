package dnsmanager

import "expvar"

var DNSMetrics = struct {
	LookupAddrAttempts *expvar.Int
	LookupAddrSuccess  *expvar.Int
	LookupAddrErrors   *expvar.Int
	CacheMisses        *expvar.Int
}{
	LookupAddrAttempts: expvar.NewInt("dns_lookup_addr_attempts"),
	LookupAddrSuccess:  expvar.NewInt("dns_lookup_addr_success"),
	LookupAddrErrors:   expvar.NewInt("dns_lookup_addr_errors"),
	CacheMisses:        expvar.NewInt("dns_cache_misses"),
}

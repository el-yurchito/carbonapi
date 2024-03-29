package dnsmanager

import (
	"context"
	"sync"
	"time"

	"github.com/rs/dnscache"
)

type DNSManager struct {
	r *dnscache.Resolver
}

const (
	unknownRecordSign string = "unknown"

	resolveTimeout time.Duration = 10 * time.Second
	refreshTimeout time.Duration = 3600 * time.Second
)

var (
	once sync.Once
	dm   *DNSManager
)

func Get() *DNSManager {
	once.Do(func() {
		dm = &DNSManager{
			r: &dnscache.Resolver{
				Timeout:     resolveTimeout,
				OnCacheMiss: func() { DNSMetrics.CacheMisses.Add(1) },
			},
		}
		go func() {
			for {
				time.Sleep(refreshTimeout)
				dm.r.RefreshWithOptions(dnscache.ResolverRefreshOptions{
					ClearUnused:      true,
					PersistOnFailure: true,
				})
			}
		}()
	})
	return dm
}

func (dm *DNSManager) GetDomainNameByIP(ip string) string {
	DNSMetrics.LookupAddrAttempts.Add(1)
	cachedRecord, cacheErr := dm.r.LookupAddr(context.Background(), ip)
	if cacheErr != nil || len(cachedRecord) == 0 {
		DNSMetrics.LookupAddrErrors.Add(1)
		return unknownRecordSign
	} else {
		DNSMetrics.LookupAddrSuccess.Add(1)
		return cachedRecord[0]
	}
}

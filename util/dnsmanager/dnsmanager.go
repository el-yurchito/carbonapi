package dnsmanager

import (
	"sync"

	"github.com/go-graphite/carbonzipper/cache"
)

type DNSManager struct {
	dnsCache cache.BytesCache
}

const (
	cacheMaxSizeBytes       uint64 = 100 * 1024 * 1024 // 100 MB
	recordExpirationSeconds int32  = 3600              // 1 hour
	unknownRecordSign       string = "unknown"         // In case when of unsuccessful lookup
)

var (
	once sync.Once
	dm   *DNSManager
)

func Get() *DNSManager {
	once.Do(func() {
		dm = &DNSManager{
			dnsCache: cache.NewExpireCache(cacheMaxSizeBytes),
		}
	})
	return dm
}

func (dm *DNSManager) GetDomainNameByIP(ip string) string {
	cachedRecord, cacheErr := dm.dnsCache.Get(ip)
	if cacheErr != nil {
		name := getDomainNameByIP(ip)
		if name != unknownRecordSign {
			dm.dnsCache.Set(ip, []byte(name), recordExpirationSeconds)
		}
		return name
	} else {
		return string(cachedRecord)
	}
}

package dnsmanager

import (
	"net"
)

func getDomainNameByIP(ip string) string {
	names, lookupErr := net.LookupAddr(ip)
	if lookupErr != nil || len(names) == 0 {
		return unknownRecordSign
	}
	return names[0]
}

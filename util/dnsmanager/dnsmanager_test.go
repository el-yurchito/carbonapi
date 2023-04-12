package dnsmanager

import (
	"testing"
)

func TestGetDomainNameByIPUnknown(t *testing.T) {
	dm := Get()
	unknownDN := dm.GetDomainNameByIP("unknown")
	if unknownDN != "unknown" {
		t.Errorf("could not resolve localhost IP address; returned value: %s", unknownDN)
	}
}

func TestGetDomainNameByIPEmpty(t *testing.T) {
	dm := Get()
	emptyDN := dm.GetDomainNameByIP("")
	if emptyDN != "unknown" {
		t.Errorf("could not resolve localhost IP address; returned value: %s", emptyDN)
	}
}

func TestGetDomainNameByIPLocalhost(t *testing.T) {
	dm := Get()
	localhostDN := dm.GetDomainNameByIP("127.0.0.1")
	if localhostDN != "localhost" {
		t.Errorf("could not resolve localhost IP address; returned value: %s", localhostDN)
	}
}

func TestGetDomainNameByIPGoogle(t *testing.T) {
	dm := Get()
	googleDNS := dm.GetDomainNameByIP("8.8.8.8")
	if googleDNS != "dns.google." {
		t.Errorf("could not resolve localhost IP address; returned value: %s", googleDNS)
	}
}

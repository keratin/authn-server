package config

import (
	"net/url"
	"strings"
)

type Domain struct {
	Hostname string
	Port     string
}

func ParseDomain(domain string) Domain {
	pieces := strings.SplitN(domain, ":", 2)
	if len(pieces) == 1 {
		pieces = append(pieces, "")
	}
	return Domain{Hostname: pieces[0], Port: pieces[1]}
}

func (d *Domain) Matches(origin *url.URL) bool {
	// hostname must always match.
	if d.Hostname != origin.Hostname() {
		return false
	}

	// if port is specified, it must match.
	if d.Port == "" {
		return true
	}
	originPort := origin.Port()
	if originPort == "" && origin.Scheme == "http" {
		originPort = "80"
	}
	if originPort == "" && origin.Scheme == "https" {
		originPort = "443"
	}
	if d.Port == originPort {
		return true
	}
	return false
}

func (d *Domain) String() string {
	if d.Port == "" {
		return d.Hostname
	}
	return d.Hostname + ":" + d.Port
}

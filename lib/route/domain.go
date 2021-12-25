package route

import (
	"net/url"
	"strings"
)

// Domain is subset of url.URL that enables a fuzzy match. A Domain must always have a Hostname, and
// may also have a Port.
type Domain struct {
	Hostname string
	Port     string
}

// ParseDomain will parse a string containing either host or host:port and return a Domain.
func ParseDomain(domain string) Domain {
	pieces := strings.SplitN(domain, ":", 2)
	if len(pieces) == 1 {
		pieces = append(pieces, "")
	}
	return Domain{Hostname: pieces[0], Port: pieces[1]}
}

// FindDomain returns a matching domain if the given string is a URL that matches
func FindDomain(str string, domains []Domain) *Domain {
	originURL, err := url.Parse(str)
	if err != nil {
		return nil
	}

	for _, d := range domains {
		if d.Matches(originURL) {
			return &d
		}
	}
	return nil
}

// Matches will compare the Domain against a given URL. The Hostname must always be a perfect match,
// and if Port is specified (non-blank) then it must also match. The common ports 80 and 443 will be
// satisfied by http and https schemes, respectively.
func (d *Domain) Matches(origin *url.URL) bool {
	// Hostname must match.
	if !Match(d.Hostname, origin.Hostname()) {
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

// String converts a Domain back into a host or host:port string.
func (d *Domain) String() string {
	if d.Port == "" {
		return d.Hostname
	}
	return d.Hostname + ":" + d.Port
}

// URL converts a Domain into a URL
func (d *Domain) URL() url.URL {
	if d.Port == "80" {
		return url.URL{Scheme: "http", Host: d.Hostname}
	}
	if d.Port == "443" {
		return url.URL{Scheme: "https", Host: d.Hostname}
	}
	return url.URL{Scheme: "http", Host: d.String()}
}

// Match returns true if the text satisfies the pattern
// string, supporting only '*' wildcard in the pattern.
func Match(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	return deepMatchRune([]rune(name), []rune(pattern))
}

func deepMatchRune(str, pattern []rune) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		default:
			if len(str) == 0 || str[0] != pattern[0] {
				return false
			}
		case '*':
			return deepMatchRune(str, pattern[1:]) ||
				(len(str) > 0 && deepMatchRune(str[1:], pattern))
		}

		str = str[1:]
		pattern = pattern[1:]
	}

	return len(str) == 0 && len(pattern) == 0
}

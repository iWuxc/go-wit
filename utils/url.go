package utils

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

// AddHTTPScheme .
func AddHTTPScheme(endpoints []string) []string {
	for index := range endpoints {
		urlIns := url.URL{}
		urlIns.Scheme = "http"
		urlIns.Host = endpoints[index]

		// set endpoint.
		endpoints[index] = urlIns.String()
	}
	return endpoints
}

// AddHTTPSScheme .
func AddHTTPSScheme(endpoints []string) []string {
	for index := range endpoints {
		urlIns := url.URL{}
		urlIns.Scheme = "https"
		urlIns.Host = endpoints[index]

		// set endpoint.
		endpoints[index] = urlIns.String()
	}
	return endpoints
}

// MakeURL .
func MakeURL(url2 url.URL) string {
	if len(url2.Scheme) == 0 {
		url2.Scheme = "https"
	}

	return url2.String()
}

// NewEndpoint new an Endpoint URL.
func NewEndpoint(scheme, host string) *url.URL {
	return &url.URL{Scheme: scheme, Host: host}
}

// Scheme is the scheme of endpoint url.
func Scheme(scheme string, isSecure bool) string {
	if isSecure {
		return scheme + "s"
	}
	return scheme
}

// Extract returns a private addr and port.
func Extract(hostPort string, lis net.Listener) (string, error) {
	addr, port, err := net.SplitHostPort(hostPort)
	if err != nil && lis == nil {
		return "", err
	}
	if lis != nil {
		p, ok := Port(lis)
		if !ok {
			return "", fmt.Errorf("failed to extract port: %v", lis.Addr())
		}
		port = strconv.Itoa(p)
	}
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		return net.JoinHostPort(addr, port), nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	lowest := int(^uint(0) >> 1)
	var result net.IP
	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		if iface.Index < lowest || result == nil {
			lowest = iface.Index
		}
		if result != nil {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, rawAddr := range addrs {
			var ip net.IP
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}
			if isValidIP(ip.String()) {
				result = ip
			}
		}
	}
	if result != nil {
		return net.JoinHostPort(result.String(), port), nil
	}
	return "", nil
}

// Port return a real port.
func Port(lis net.Listener) (int, bool) {
	if addr, ok := lis.Addr().(*net.TCPAddr); ok {
		return addr.Port, true
	}
	return 0, false
}

func isValidIP(addr string) bool {
	ip := net.ParseIP(addr)
	return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
}
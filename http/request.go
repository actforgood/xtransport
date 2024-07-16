package http

import (
	"io"
	"net"
	"net/http"
	"strings"
)

const defaultMaxBodyBytes int64 = 524288 // 0.5 Mb

// GetRequestBody MaxBytesReader prevents clients from accidentally or maliciously
// sending a large request and wasting server resources.
func GetRequestBody(w http.ResponseWriter, r *http.Request, maxBodyBytes ...int64) io.Reader {
	var maxBody int64
	if nil != maxBodyBytes && maxBodyBytes[0] > 0 {
		maxBody = maxBodyBytes[0]
	} else {
		maxBody = defaultMaxBodyBytes
	}

	return http.MaxBytesReader(w, r.Body, maxBody)
}

// GetClientIP returns client ip from following sources, in this order:
// X-Real-Ip, X-Forwarded-For (first ip in an eventual list), r.RemoteAddr.
// Note: we assume these headers come from a trusted proxy,
// as client can also set them to any arbitrary value which may lead to ip spoofing.
func GetClientIP(r *http.Request) net.IP {
	if r.Header.Get("X-Real-Ip") != "" {
		if ip := net.ParseIP(r.Header.Get("X-Real-Ip")); ip != nil {
			return ip
		}
	}
	if r.Header.Get("X-Forwarded-For") != "" {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		for _, ip := range ips {
			if ip := net.ParseIP(strings.TrimSpace(ip)); ip != nil {
				return ip
			}
		}
	}
	addr, _, _ := net.SplitHostPort(r.RemoteAddr)

	return net.ParseIP(addr)
}

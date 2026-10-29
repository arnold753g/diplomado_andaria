package middleware

import (
	"context"
	"net"
	"net/http"
	"net/netip"
)

const trustedProxyKey contextKey = "trusted-proxy"

func TrustedProxies(enabled bool, cidrs []string) func(http.Handler) http.Handler {
	var networks []netip.Prefix
	for _, cidr := range cidrs {
		if network, err := netip.ParsePrefix(cidr); err == nil && network.Bits() > 0 {
			networks = append(networks, network)
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			peer, err := netip.ParseAddr(host)
			trusted := false
			if enabled && err == nil {
				for _, network := range networks {
					if network.Contains(peer.Unmap()) {
						trusted = true
						break
					}
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), trustedProxyKey, trusted)))
		})
	}
}

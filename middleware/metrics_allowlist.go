package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseMetricsAllowlistNets parses CIDR strings into nets; empty input returns nil, nil.
func ParseMetricsAllowlistNets(cidrs []string) ([]*net.IPNet, error) {
	if len(cidrs) == 0 {
		return nil, nil
	}
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, raw := range cidrs {
		c := strings.TrimSpace(raw)
		if c == "" {
			return nil, fmt.Errorf("empty CIDR in metrics allowlist")
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, fmt.Errorf("parse CIDR %q: %w", c, err)
		}
		out = append(out, n)
	}
	return out, nil
}

func peerIP(r *http.Request) net.IP {
	if r == nil {
		return nil
	}
	addr := strings.TrimSpace(r.RemoteAddr)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.ParseIP(addr)
	}
	return net.ParseIP(host)
}

// MetricsAllowlist 限制仅 metricsPath 且客户端 TCP 源 IP 命中 allow 时才进入后续 handler（与 X-Forwarded-For 无关，避免伪造）。
func MetricsAllowlist(metricsPath string, allow []*net.IPNet) gin.HandlerFunc {
	if len(allow) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	return func(c *gin.Context) {
		if c.Request.URL.Path != metricsPath {
			c.Next()
			return
		}
		ip := peerIP(c.Request)
		if ip == nil || !ipInAnyNet(ip, allow) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

func ipInAnyNet(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

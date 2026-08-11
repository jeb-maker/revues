// Package safehttp provides SSRF-resistant HTTP clients for outbound integrations.
package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

// Options configure outbound SSRF policy.
type Options struct {
	// AllowDevLocalhost permits http://localhost and dialing loopback (dev only).
	AllowDevLocalhost bool
	// Timeout is the overall client timeout.
	Timeout time.Duration
	// DialTimeout bounds DNS+TCP connect.
	DialTimeout time.Duration
	// MaxRedirects caps followed redirects (default 1).
	MaxRedirects int
}

// NewClient returns an HTTP client that resolves hosts once, rejects blocked IPs,
// and dials the validated IP (no second DNS lookup — mitigates rebinding).
func NewClient(opts Options) *http.Client {
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = opts.Timeout
	}
	if opts.MaxRedirects <= 0 {
		opts.MaxRedirects = 1
	}
	return &http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= opts.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			if err := ValidateURL(req.URL.String(), opts.AllowDevLocalhost); err != nil {
				return err
			}
			_, err := ResolveAllowedIP(req.Context(), req.URL.Hostname(), opts.AllowDevLocalhost)
			return err
		},
		Transport: &http.Transport{
			DialContext: pinnedDialContext(opts),
		},
	}
}

func pinnedDialContext(opts Options) func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: opts.DialTimeout}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("split dial addr: %w", err)
		}
		ip, err := ResolveAllowedIP(ctx, host, opts.AllowDevLocalhost)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}
}

// ValidateURL checks scheme policy (https, or http+localhost in dev).
func ValidateURL(raw string, allowDevLocalhost bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "https" {
		return nil
	}
	if allowDevLocalhost && u.Scheme == "http" && IsLocalhostHost(u.Hostname()) {
		return nil
	}
	return fmt.Errorf("url scheme not allowed")
}

// ResolveAllowedIP resolves host and returns the first non-blocked IP.
func ResolveAllowedIP(ctx context.Context, host string, allowDevLocalhost bool) (net.IP, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	if IsLocalhostHost(host) {
		if allowDevLocalhost {
			return net.ParseIP("127.0.0.1"), nil
		}
		return nil, fmt.Errorf("localhost not allowed")
	}
	// Literal IP in the URL/host — validate without DNS.
	if ip := net.ParseIP(host); ip != nil {
		if BlockedIP(ip) {
			return nil, fmt.Errorf("blocked ip address")
		}
		return ip, nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve host: %w", err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("resolve host: no addresses")
	}
	for _, ipAddr := range ips {
		if BlockedIP(ipAddr.IP) {
			return nil, fmt.Errorf("blocked ip address")
		}
	}
	// Prefer IPv4 for JoinHostPort simplicity when available.
	for _, ipAddr := range ips {
		if v4 := ipAddr.IP.To4(); v4 != nil {
			return v4, nil
		}
	}
	return ips[0].IP, nil
}

// BlockedIP reports whether ip must not be dialed (private, loopback, link-local, CGNAT, metadata).
func BlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	addr = addr.Unmap()
	// CGNAT 100.64.0.0/10
	if addr.Is4() {
		a := addr.As4()
		if a[0] == 100 && a[1] >= 64 && a[1] <= 127 {
			return true
		}
	}
	if addr.String() == "169.254.169.254" {
		return true
	}
	return false
}

// IsLocalhostHost reports whether host is a loopback name or address.
func IsLocalhostHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

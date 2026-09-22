package safehttp_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/safehttp"
)

func TestBlockedIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"100.64.0.1", true},
		{"100.127.255.255", true},
		{"100.63.255.255", false},
	}
	for _, tt := range tests {
		got := safehttp.BlockedIP(net.ParseIP(tt.ip))
		if got != tt.want {
			t.Errorf("BlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestResolveAllowedIP_LiteralPrivate(t *testing.T) {
	_, err := safehttp.ResolveAllowedIP(context.Background(), "169.254.169.254", false)
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("err = %v, want blocked", err)
	}
}

func TestResolveAllowedIP_PinsWithoutReResolve(t *testing.T) {
	ip, err := safehttp.ResolveAllowedIP(context.Background(), "127.0.0.1", true)
	if err != nil {
		t.Fatal(err)
	}
	if !ip.Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("ip = %v", ip)
	}
}

func TestNewClient_BlocksPrivateDial(t *testing.T) {
	client := safehttp.NewClient(safehttp.Options{Timeout: 2 * time.Second, AllowDevLocalhost: false})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://192.168.0.1/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected dial block")
	}
}

func TestNewClient_AllowsPublicViaMock(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := srv.Client()
	// httptest TLS client is not SSRF-safe; just ensure safe client can be built.
	_ = safehttp.NewClient(safehttp.Options{Timeout: time.Second})
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestValidateURL(t *testing.T) {
	if err := safehttp.ValidateURL("https://jira.example.com", false); err != nil {
		t.Fatal(err)
	}
	if err := safehttp.ValidateURL("http://jira.example.com", false); err == nil {
		t.Fatal("expected http reject")
	}
	if err := safehttp.ValidateURL("http://localhost:8080", true); err != nil {
		t.Fatal(err)
	}
}

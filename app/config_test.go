package app

import "testing"

func TestResolveSSEURL(t *testing.T) {
	tests := []struct {
		name     string
		explicit string
		appURL   string
		ssePort  string
		want     string
	}{
		{"explicit wins over derivation", "https://events.acme.com", "http://localhost:8080", "8090", "https://events.acme.com"},
		{"explicit trailing slash trimmed", "https://events.acme.com/", "http://localhost:8080", "8090", "https://events.acme.com"},
		{"derives from app url host", "", "http://localhost:8080", "8090", "http://localhost:8090"},
		{"derives across a LAN address", "", "http://192.168.100.250:8080", "8090", "http://192.168.100.250:8090"},
		{"preserves https scheme", "", "https://acme.com", "8090", "https://acme.com:8090"},
		{"app url without a port", "", "http://localhost", "8090", "http://localhost:8090"},
		{"scheme-less app url falls back to http", "", "localhost:8080", "8090", "http://localhost:8090"},
		{"protocol-relative app url gets http", "", "//localhost:8080", "8090", "http://localhost:8090"},
		{"ipv6 host is bracketed", "", "http://[::1]:8080", "8090", "http://[::1]:8090"},
		{"empty app url falls back to localhost", "", "", "8090", "http://localhost:8090"},
		{"honors a non-default sse port", "", "http://localhost:8080", "9999", "http://localhost:9999"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := resolveSSEURL(test.explicit, test.appURL, test.ssePort)
			if got != test.want {
				t.Errorf("resolveSSEURL(%q, %q, %q) = %q, want %q", test.explicit, test.appURL, test.ssePort, got, test.want)
			}
		})
	}
}

package auth

import (
	"net/http"
	"testing"
)

func TestClientIPBehindProxyHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "direct public peer ignores spoofed xff",
			remoteAddr: "203.0.113.50:44321",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.9"},
			want:       "203.0.113.50",
		},
		{
			name:       "loopback peer trusts cf-connecting-ip over xff",
			remoteAddr: "127.0.0.1:54321",
			headers: map[string]string{
				"CF-Connecting-IP": "198.51.100.20",
				"X-Forwarded-For":  "203.0.113.1, 172.16.0.2",
				"X-Real-IP":        "172.16.0.2",
			},
			want: "198.51.100.20",
		},
		{
			name:       "docker bridge peer trusts true-client-ip",
			remoteAddr: "172.18.0.1:44000",
			headers: map[string]string{
				"True-Client-IP":  "198.51.100.30",
				"X-Forwarded-For": "203.0.113.2",
			},
			want: "198.51.100.30",
		},
		{
			name:       "nginx x-real-ip when no cdn header",
			remoteAddr: "10.0.0.2:80",
			headers: map[string]string{
				"X-Real-IP":       "198.51.100.40",
				"X-Forwarded-For": "198.51.100.40, 10.0.0.2",
			},
			want: "198.51.100.40",
		},
		{
			name:       "xff leftmost valid hop",
			remoteAddr: "192.168.1.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.50, 172.16.0.5, 192.168.1.1",
			},
			want: "198.51.100.50",
		},
		{
			name:       "xff skips garbage then returns first ip",
			remoteAddr: "10.0.0.5:1",
			headers: map[string]string{
				"X-Forwarded-For": "unknown, 198.51.100.60",
			},
			want: "198.51.100.60",
		},
		{
			name:       "private peer without headers uses remote",
			remoteAddr: "10.0.0.8:9999",
			want:       "10.0.0.8",
		},
		{
			name:       "ipv6 loopback trusts cf",
			remoteAddr: "[::1]:443",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			want: "2001:db8::1",
		},
		{
			name:       "cgnat peer trusts xff",
			remoteAddr: "100.64.0.10:80",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.70",
			},
			want: "198.51.100.70",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			got := clientIP(req)
			if got != tc.want {
				t.Fatalf("clientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

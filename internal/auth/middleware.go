package auth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/replacearr/replacearr/internal/domain"
)

type contextKey string

const (
	userContextKey    contextKey = "replacearr_user"
	apiKeyContextKey  contextKey = "replacearr_apikey"
	sessionContextKey contextKey = "replacearr_session"
)

func UserFromContext(ctx context.Context) *domain.User {
	if ctx == nil {
		return nil
	}
	u, _ := ctx.Value(userContextKey).(*domain.User)
	return u
}

func APIKeyFromContext(ctx context.Context) *domain.APIKey {
	if ctx == nil {
		return nil
	}
	k, _ := ctx.Value(apiKeyContextKey).(*domain.APIKey)
	return k
}

func SessionFromContext(ctx context.Context) *domain.AuthSession {
	if ctx == nil {
		return nil
	}
	sess, _ := ctx.Value(sessionContextKey).(*domain.AuthSession)
	return sess
}

func (s *Service) Middleware(requiredScope string, requiredPerm string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user    *domain.User
			apiKey  *domain.APIKey
			session *domain.AuthSession
		)

		// 1. X-Api-Key header
		if hdrKey := strings.TrimSpace(r.Header.Get("X-Api-Key")); hdrKey != "" {
			if k, u, err := s.ValidateAPIKey(hdrKey, requiredScope); err == nil {
				apiKey = k
				user = u
			}
		} else if qKey := extractQueryAPIKey(r); qKey != "" {
			// 2. ?apikey= or ?apiKey= query parameter
			if k, u, err := s.ValidateAPIKey(qKey, requiredScope); err == nil {
				apiKey = k
				user = u
			}
		} else if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			// 3. Authorization: Bearer <token> (try JWT first, then API key)
			token := strings.TrimSpace(authHeader[7:])
			if token != "" {
				if u, sess, err := s.ValidateToken(token, clientIP(r)); err == nil {
					user = u
					session = sess
				} else if k, u, err := s.ValidateAPIKey(token, requiredScope); err == nil {
					apiKey = k
					user = u
				}
			}
		} else if cookie, err := r.Cookie("replacearr_token"); err == nil && strings.TrimSpace(cookie.Value) != "" {
			// 4. replacearr_token cookie
			if u, sess, err := s.ValidateToken(cookie.Value, clientIP(r)); err == nil {
				user = u
				session = sess
			}
		}

		if user == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		if requiredPerm != "" && !s.HasPermission(user, apiKey, requiredPerm) {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "forbidden: missing permission " + requiredPerm,
			})
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		if apiKey != nil {
			ctx = context.WithValue(ctx, apiKeyContextKey, apiKey)
		}
		if session != nil {
			ctx = context.WithValue(ctx, sessionContextKey, session)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) RequirePermission(perm string, next http.HandlerFunc) http.HandlerFunc {
	return s.Middleware("", perm, next).ServeHTTP
}

func extractQueryAPIKey(r *http.Request) string {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("apikey")); v != "" {
		return v
	}
	if v := strings.TrimSpace(q.Get("apiKey")); v != "" {
		return v
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func extractBearerOrCookieToken(r *http.Request) string {
	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	if cookie, err := r.Cookie("replacearr_token"); err == nil {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

// clientIP resolves the originating client address.
//
// When the immediate peer looks like a reverse proxy (loopback / private /
// link-local — the usual nginx, Traefik, Caddy, Cloudflare Tunnel, or Docker
// hop), forwarding headers are consulted in CDN-first order:
//
//  1. CF-Connecting-IP (Cloudflare)
//  2. True-Client-IP (Cloudflare Enterprise / some CDNs)
//  3. X-Real-IP (nginx)
//  4. X-Forwarded-For (left-most valid hop — original client)
//
// Headers are ignored when the peer is a public address so a client talking
// directly to Replacearr cannot spoof its IP. Values that are not parseable
// IPs are skipped.
func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	remote := remoteIP(r.RemoteAddr)
	if remote != "" && !isTrustedProxy(remote) {
		return remote
	}
	if ip := headerClientIP(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := headerClientIP(r.Header.Get("True-Client-IP")); ip != "" {
		return ip
	}
	if ip := headerClientIP(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			if ip := headerClientIP(part); ip != "" {
				return ip
			}
		}
	}
	return remote
}

func remoteIP(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Bare IP without port (or already stripped).
		if ip := net.ParseIP(addr); ip != nil {
			return ip.String()
		}
		return addr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return host
}

func headerClientIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// Some proxies wrap IPv6 as "[addr]".
	raw = strings.Trim(raw, "[]")
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	return ip.String()
}

func isTrustedProxy(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	// CGNAT / shared-address space often used between reverse proxies and app
	// containers (RFC 6598).
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
		return true
	}
	return false
}

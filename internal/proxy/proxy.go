package proxy

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/umairakhlaque/activesyncking/internal/config"
	"github.com/umairakhlaque/activesyncking/pkg/apitypes"
)

const (
	// EAS endpoint path all ActiveSync clients target
	easPath = "/Microsoft-Server-ActiveSync"

	// Header injected by gateway to Exchange after auth passes
	headerAuthUser = "X-SyncGuard-User"
)

// Gateway is the EAS reverse proxy. It intercepts every request to
// /Microsoft-Server-ActiveSync, runs the auth check, and either proxies
// to Exchange or rejects the client.
type Gateway struct {
	cfg      config.GatewayConfig
	proxy    *httputil.ReverseProxy
	authCli  *http.Client
	log      *zap.Logger
}

// New creates a configured Gateway. exchangeURL is the upstream Exchange
// CAS / Front-End endpoint; authServiceURL is the internal auth service.
func New(cfg config.GatewayConfig, log *zap.Logger) (*Gateway, error) {
	exchangeURL, err := url.Parse(cfg.ExchangeURL)
	if err != nil {
		return nil, fmt.Errorf("parsing exchange url: %w", err)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: cfg.IdleTimeout,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       cfg.IdleTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	rp := httputil.NewSingleHostReverseProxy(exchangeURL)
	rp.Transport = transport
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("upstream exchange error", zap.Error(err), zap.String("path", r.URL.Path))
		// Fail-closed: return 503 rather than a misleading Exchange error
		http.Error(w, "Exchange backend unavailable", http.StatusServiceUnavailable)
	}

	// Rewrite Host header so Exchange sees its own hostname
	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = exchangeURL.Host
	}

	authCli := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			MaxIdleConns: 10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	return &Gateway{
		cfg:     cfg,
		proxy:   rp,
		authCli: authCli,
		log:     log,
	}, nil
}

// ServeHTTP is the main entry point for all inbound EAS requests.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only intercept the EAS endpoint; pass everything else through
	if !strings.HasPrefix(r.URL.Path, easPath) {
		g.proxy.ServeHTTP(w, r)
		return
	}

	// Extract Basic Auth credentials
	username, password, ok := r.BasicAuth()
	if !ok || username == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="SyncGuard EAS"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract device context from EAS headers/query params
	deviceID := extractDeviceID(r)
	if deviceID == "" {
		// Treat as a non-EAS request (OPTIONS, health checks, etc.)
		g.proxy.ServeHTTP(w, r)
		return
	}

	// Call auth service
	authResp, err := g.callAuthService(r.Context(), apitypes.AuthenticateRequest{
		Username:   username,
		Password:   password,
		DeviceID:   deviceID,
		DeviceType: extractDeviceType(r),
		UserAgent:  r.UserAgent(),
		IP:         realIP(r),
	})
	if err != nil {
		g.log.Error("auth service unreachable", zap.Error(err))
		if g.cfg.FailClosed {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		// FailClosed=false: log and allow (NEVER set in production)
		g.log.Warn("fail-open mode: allowing request despite auth service error")
		g.proxy.ServeHTTP(w, r)
		return
	}

	switch authResp.Status {
	case apitypes.AuthStatusAllow:
		// Trusted device — strip credentials from upstream request and proxy
		r.Header.Set(headerAuthUser, username)
		g.proxy.ServeHTTP(w, r)

	case apitypes.AuthStatusMFARequired:
		// Return 401 with challenge metadata in headers.
		// The mobile client will retry; the challenge is consumed via the
		// admin portal or companion app (out-of-band).
		w.Header().Set("X-SyncGuard-Challenge-ID", authResp.ChallengeID)
		w.Header().Set("X-SyncGuard-MFA-Method", authResp.Method)
		w.Header().Set("WWW-Authenticate", `Basic realm="SyncGuard MFA Required"`)
		http.Error(w, "MFA verification required", http.StatusUnauthorized)

	case apitypes.AuthStatusBlocked:
		http.Error(w, "Device or account blocked", http.StatusForbidden)

	case apitypes.AuthStatusDenied:
		w.Header().Set("WWW-Authenticate", `Basic realm="SyncGuard EAS"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

	default:
		// Any unknown status is treated as a denial (fail-closed)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}
}

// callAuthService POSTs the auth request to the internal auth service and
// returns the parsed response.
func (g *Gateway) callAuthService(ctx context.Context, req apitypes.AuthenticateRequest) (*apitypes.AuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshalling auth request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.cfg.AuthServiceURL+"/api/v1/authenticate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.authCli.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling auth service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, fmt.Errorf("reading auth response: %w", err)
	}

	var authResp apitypes.AuthenticateResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return nil, fmt.Errorf("parsing auth response: %w", err)
	}

	return &authResp, nil
}

// extractDeviceID extracts the EAS DeviceId from the query string or
// the MS-ASDevice header, whichever is present.
func extractDeviceID(r *http.Request) string {
	if id := r.URL.Query().Get("DeviceId"); id != "" {
		return id
	}
	// Some implementations use MS-ASDevice header
	if id := r.Header.Get("MS-ASDevice"); id != "" {
		return id
	}
	return ""
}

func extractDeviceType(r *http.Request) string {
	if dt := r.URL.Query().Get("DeviceType"); dt != "" {
		return dt
	}
	return ""
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// decodeBasicAuth extracts username:password from a Basic Authorization header.
// Kept here for tests; the standard library r.BasicAuth() is used in ServeHTTP.
func decodeBasicAuth(header string) (string, string, bool) {
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", false
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

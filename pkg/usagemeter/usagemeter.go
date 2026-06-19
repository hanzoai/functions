// Package usagemeter wires the Functions router into Hanzo commerce billing so
// that function invocations are paid for — the same way the LLM/cloud path is.
//
// It is a thin, env-gated adapter over the shared metering hook
// (github.com/hanzoai/commerce/metering): one reusable client, no per-product
// reimplementation. When commerce is not configured (no COMMERCE_URL, or
// METERING_DISABLED=true) the middleware is a transparent pass-through, so this
// is safe to wire unconditionally and turns on only when the operator injects
// the (KMS-backed) commerce service token.
//
// Wiring (router.go):
//
//	mux.Use(metrics.HTTPMetricMiddleware)
//	mux.Use(usagemeter.Middleware(logger)) // gate + record per invocation
package usagemeter

import (
	"net/http"
	"os"
	"strconv"

	"github.com/hanzoai/commerce/metering"
	"go.uber.org/zap"
)

// Provider label recorded on commerce usage transactions for this service.
const Provider = "functions"

// EnvPricePerCallCents sets the flat per-successful-invocation price in cents.
// Default 0 — i.e. gate-only until a price is configured (so enabling billing
// in two steps, gate then charge, is possible). A real deployment sets this
// (or replaces PriceFunc) to charge per invocation.
const EnvPricePerCallCents = "FUNCTIONS_PRICE_PER_CALL_CENTS"

// Paths that must never be gated or charged — infra endpoints.
var skipPaths = map[string]bool{
	"/router-healthz": true,
	"/metrics":        true,
}

// Middleware returns gorilla/mux-compatible middleware that, per request:
//   - gates on the caller's commerce balance (fail-closed by default), and
//   - records usage to commerce after a successful invocation.
//
// Identity comes from the gateway-minted X-User-Id / X-Org-Id headers (the
// trust boundary). The returned middleware has the standard
// func(http.Handler) http.Handler shape, so `mux.Use(...)` accepts it directly.
func Middleware(logger *zap.Logger) func(http.Handler) http.Handler {
	meter, err := metering.FromEnv()
	if err != nil {
		// Misconfigured URL — log and fall back to pass-through rather than
		// crash the router. A bad COMMERCE_URL must not take Functions down.
		if logger != nil {
			logger.Error("usagemeter: invalid commerce config, metering disabled", zap.Error(err))
		}
		return passthrough
	}

	if !meter.Enabled() {
		if logger != nil {
			logger.Info("usagemeter: commerce not configured, metering disabled (pass-through)")
		}
		return passthrough
	}

	price := priceFromEnv()
	if logger != nil {
		logger.Info("usagemeter: metering enabled", zap.Int64("pricePerCallCents", price))
	}

	return meter.Middleware(metering.MiddlewareConfig{
		Provider: Provider,
		Price: func(_ *http.Request, status int, _ metering.AuthInput) int64 {
			if status >= 200 && status < 400 {
				return price
			}
			return 0 // never charge for failed invocations
		},
		Skip: func(r *http.Request) bool { return skipPaths[r.URL.Path] },
		OnRecordError: func(r *http.Request, u metering.Usage, e error) {
			if logger != nil {
				logger.Warn("usagemeter: failed to record usage",
					zap.String("user", u.User), zap.Int64("amountCents", u.AmountCents), zap.Error(e))
			}
		},
	})
}

func priceFromEnv() int64 {
	if v := os.Getenv(EnvPricePerCallCents); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			return n
		}
	}
	return 0
}

func passthrough(next http.Handler) http.Handler { return next }

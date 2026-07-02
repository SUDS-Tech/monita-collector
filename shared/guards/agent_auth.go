package guards

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/bastion-framework/bast"

	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
)

// AgentInfo is set in request context by AgentAuthGuard on successful auth.
// Key: "agent". Handlers read it via ctx.MustGet("agent").(*guards.AgentInfo).
type AgentInfo struct {
	ID               string
	OrgID            string
	SigningKeyHash   string // raw signing key (used as HMAC key)
	FingerprintHash  string // stored hash — included in HMAC signed material
	RotationRequired bool
	Frozen           bool
	Revoked          bool
	ExpiresAt        time.Time
}

// AgentResolver is implemented by agents.Service (duck typing).
type AgentResolver interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*AgentInfo, error)
	UpdateLastSeen(ctx context.Context, agentID string) error
}

type AgentAuthGuard struct {
	agents AgentResolver
	nonce  NonceCache
}

func NewAgentAuth(agents AgentResolver, nonce NonceCache) *AgentAuthGuard {
	return &AgentAuthGuard{agents: agents, nonce: nonce}
}

func (g *AgentAuthGuard) SecurityScheme() bast.SecurityScheme {
	return bast.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "AgentToken",
		Description:  "Agent bearer token with HMAC-SHA256 request signing (X-Timestamp, X-Nonce, X-Signature, X-Fingerprint).",
	}
}

// Check implements PROTOCOL.md §2.4 agent authentication.
func (g *AgentAuthGuard) Check(ctx *bast.Ctx) error {
	// 1. Bearer token → SHA-256 hash → look up agent
	header := ctx.Header("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return bast.ErrUnauthorized(apperr.CodeAgentUnknown, "missing bearer token")
	}
	rawToken := strings.TrimPrefix(header, "Bearer ")
	tokenHash := sha256Hex(rawToken)

	agent, err := g.agents.FindByTokenHash(ctx.Context(), tokenHash)
	if err != nil {
		return bast.ErrUnauthorized(apperr.CodeAgentUnknown, "agent not found")
	}

	// 2. State checks
	if agent.Revoked {
		return bast.ErrUnauthorized(apperr.CodeAgentRevoked, "agent revoked")
	}
	if agent.Frozen {
		return bast.ErrUnauthorized(apperr.CodeAgentFrozen, "agent frozen")
	}
	if time.Now().After(agent.ExpiresAt) {
		return bast.ErrUnauthorized(apperr.CodeAgentExpired, "agent token expired")
	}

	// 3. Timestamp ±120 s
	tsHeader := ctx.Header("X-Timestamp")
	tsUnix, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return bast.ErrUnauthorized(apperr.CodeClockSkew, "invalid X-Timestamp")
	}
	diff := time.Since(time.Unix(tsUnix, 0))
	if diff < -120*time.Second || diff > 120*time.Second {
		return bast.ErrUnauthorized(apperr.CodeClockSkew, "timestamp out of window")
	}

	// 4. Nonce replay check
	nonce := ctx.Header("X-Nonce")
	if nonce == "" {
		return bast.ErrUnauthorized(apperr.CodeReplayDetected, "missing X-Nonce")
	}
	seen, err := g.nonce.SeenRecently(ctx.Context(), agent.ID, nonce, 120*time.Second)
	if err != nil || seen {
		return bast.ErrUnauthorized(apperr.CodeReplayDetected, "replay detected")
	}
	_ = g.nonce.Record(ctx.Context(), agent.ID, nonce, 120*time.Second)

	// 5. HMAC-SHA256 signature verification
	// signed_material = timestamp + "." + nonce + "." + fingerprint_hash + "." + sha256(body)
	// fingerprint_hash is the stored value — mismatch causes HMAC failure implicitly.
	rawBody, _ := ctx.RawBody()
	bodyHash := sha256Hex(string(rawBody))
	message := tsHeader + "." + nonce + "." + agent.FingerprintHash + "." + bodyHash

	expected := hmacHex([]byte(agent.SigningKeyHash), message)
	if !hmac.Equal([]byte(ctx.Header("X-Signature")), []byte(expected)) {
		return bast.ErrUnauthorized(apperr.CodeSignatureMismatch, "invalid signature")
	}

	// 6. Update last seen and set context
	_ = g.agents.UpdateLastSeen(ctx.Context(), agent.ID)
	ctx.Set("agent", agent)
	return nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacHex(key []byte, message string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
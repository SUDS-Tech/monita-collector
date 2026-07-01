package guards

import (
	"context"

	"github.com/bastion-framework/bast"
)

// AgentInfo is set in request context by AgentAuthGuard on successful auth.
// Key: "agent". Handlers read it via ctx.MustGet("agent").(*guards.AgentInfo).
type AgentInfo struct {
	ID              string
	OrgID           string
	SigningKeyHash   string
	FingerprintHash string
	Frozen          bool
	Revoked         bool
}

// AgentResolver is implemented by agents.Service (duck typing).
type AgentResolver interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*AgentInfo, error)
	UpdateLastSeen(ctx context.Context, agentID string) error
	FreezeAgent(ctx context.Context, agentID string) error
	SetFingerprintDrift(ctx context.Context, agentID string) error
}

type AgentAuthGuard struct {
	agents AgentResolver
	nonce  NonceCache
}

func NewAgentAuth(agents AgentResolver, nonce NonceCache) *AgentAuthGuard {
	return &AgentAuthGuard{agents: agents, nonce: nonce}
}

// Check implements bast.Guard per PROTOCOL.md §2.4.
// Full implementation is the auth module step; this stub passes all requests
// so the scaffold compiles and health/ready endpoints work immediately.
func (g *AgentAuthGuard) Check(ctx *bast.Ctx) error {
	// TODO: implement
	// 1. Parse Bearer token → SHA-256 hash → FindByTokenHash
	// 2. Validate X-Timestamp ±120 s
	// 3. X-Nonce not seen → Record
	// 4. Recompute HMAC-SHA256(signing_key, ts.nonce.fp.body_hash) → compare X-Signature
	// 5. Fingerprint tier: exact=pass, partial=drift flag, none=403+freeze+security alert
	// 6. UpdateLastSeen; ctx.Set("agent", agentInfo)
	return nil
}

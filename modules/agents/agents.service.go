package agents

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/goccy/go-json"
	"github.com/google/uuid"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
)

type Service struct {
	repo *repo
}

func newService(r *repo) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(ctx context.Context, orgID uuid.UUID, req CreateAgentRequest) (*CreateAgentResponse, error) {
	rawToken, err := generateSecret()
	if err != nil {
		return nil, err
	}
	signingKey, err := generateSecret()
	if err != nil {
		return nil, err
	}

	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.createAgent(ctx, dbsqlc.CreateAgentParams{
		OrgID:         orgID,
		Name:          req.Name,
		Hostname:      req.Hostname,
		Tags:          tagsJSON,
		TokenHash:     hashSecret(rawToken),
		SigningKeyHash: signingKey,
	})
	if err != nil {
		return nil, err
	}

	return &CreateAgentResponse{
		AgentResponse: toAgentResponse(row.ID, row.Name, row.Hostname, row.Tags, row.Frozen, row.Revoked, row.RotationRequired, row.ExpiresAt, row.LastSeenAt, row.CreatedAt),
		Token:         rawToken,
		SigningKey:    signingKey,
	}, nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]AgentResponse, error) {
	rows, err := s.repo.listAgentsByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]AgentResponse, len(rows))
	for i, r := range rows {
		out[i] = toAgentResponse(r.ID, r.Name, r.Hostname, r.Tags, r.Frozen, r.Revoked, r.RotationRequired, r.ExpiresAt, r.LastSeenAt, r.CreatedAt)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, orgID, agentID uuid.UUID) (*AgentResponse, error) {
	row, err := s.repo.getAgentByID(ctx, agentID, orgID)
	if err != nil {
		return nil, bast.ErrNotFound("AGENT_NOT_FOUND", "agent not found")
	}
	resp := toAgentResponse(row.ID, row.Name, row.Hostname, row.Tags, row.Frozen, row.Revoked, row.RotationRequired, row.ExpiresAt, row.LastSeenAt, row.CreatedAt)
	return &resp, nil
}

func (s *Service) Revoke(ctx context.Context, orgID, agentID uuid.UUID) error {
	return s.repo.revokeAgent(ctx, agentID, orgID)
}

func (s *Service) Delete(ctx context.Context, orgID, agentID uuid.UUID) error {
	return s.repo.deleteAgent(ctx, agentID, orgID)
}

// AgentResolver — implemented by duck typing for AgentAuthGuard.

func (s *Service) FindByTokenHash(ctx context.Context, tokenHash string) (*guards.AgentInfo, error) {
	row, err := s.repo.getAgentByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	fp := ""
	if row.FingerprintHash != nil {
		fp = *row.FingerprintHash
	}
	return &guards.AgentInfo{
		ID:               row.ID.String(),
		OrgID:            row.OrgID.String(),
		SigningKeyHash:   row.SigningKeyHash,
		FingerprintHash:  fp,
		RotationRequired: row.RotationRequired,
		Frozen:           row.Frozen,
		Revoked:          row.Revoked,
		ExpiresAt:        row.ExpiresAt,
	}, nil
}

func (s *Service) RotateToken(ctx context.Context, agentID string) (token, signingKey string, err error) {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return "", "", err
	}
	rawToken, err := generateSecret()
	if err != nil {
		return "", "", err
	}
	newKey, err := generateSecret()
	if err != nil {
		return "", "", err
	}
	if err := s.repo.rotateAgentToken(ctx, id, hashSecret(rawToken), newKey); err != nil {
		return "", "", err
	}
	return rawToken, newKey, nil
}

func (s *Service) UpdateLastSeen(ctx context.Context, agentID string) error {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return s.repo.updateLastSeen(ctx, id)
}

func (s *Service) FreezeAgent(ctx context.Context, agentID string) error {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return s.repo.freezeAgent(ctx, id)
}

func (s *Service) SetFingerprintDrift(ctx context.Context, agentID string) error {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return s.repo.setFingerprintDrift(ctx, id)
}

func (s *Service) SetFingerprintHash(ctx context.Context, agentID, hash string) error {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return err
	}
	return s.repo.setFingerprintHash(ctx, id, hash)
}



func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashSecret(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func toAgentResponse(id uuid.UUID, name, hostname string, tags []byte, frozen, revoked, rotationRequired bool, expiresAt time.Time, lastSeenAt **time.Time, createdAt time.Time) AgentResponse {
	return AgentResponse{
		ID:               id.String(),
		Name:             name,
		Hostname:         hostname,
		Tags:             tagsToMap(tags),
		Frozen:           frozen,
		Revoked:          revoked,
		RotationRequired: rotationRequired,
		ExpiresAt:        expiresAt,
		LastSeenAt:       derefLastSeen(lastSeenAt),
		CreatedAt:        createdAt,
	}
}
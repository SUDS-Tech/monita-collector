package logs

import (
	"context"
	"time"

	"github.com/google/uuid"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
)

type logPublisher interface {
	PublishLog(agentID, source, level, message string, lastSeen time.Time)
}

type Service struct {
	repo      *repo
	publisher logPublisher
}

func newService(r *repo, publisher logPublisher) *Service {
	return &Service{repo: r, publisher: publisher}
}

func (s *Service) Ingest(ctx context.Context, agentID uuid.UUID, req IngestRequest) (int64, error) {
	rows := make([]dbsqlc.IngestLogEntriesParams, len(req.Entries))
	for i, e := range req.Entries {
		rows[i] = dbsqlc.IngestLogEntriesParams{
			AgentID:   agentID,
			Source:    e.Source,
			Level:     e.Level,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstSeen,
			LastSeen:  e.LastSeen,
		}
	}
	n, err := s.repo.ingest(ctx, rows)
	if err != nil {
		return 0, err
	}
	if s.publisher != nil {
		for _, e := range req.Entries {
			s.publisher.PublishLog(agentID.String(), e.Source, e.Level, e.Message, e.LastSeen)
		}
	}
	return n, nil
}

func (s *Service) Query(ctx context.Context, agentID uuid.UUID, from, to time.Time, level *string, limit int32) ([]EntryResponse, error) {
	rows, err := s.repo.query(ctx, agentID, from, to, level, limit)
	if err != nil {
		return nil, err
	}
	out := make([]EntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toEntryResponse(r.AgentID, r.Source, r.Level, r.Message, r.Count, r.FirstSeen, r.LastSeen, r.ReceivedAt)
	}
	return out, nil
}

func (s *Service) Search(ctx context.Context, agentID uuid.UUID, from, to time.Time, q string, limit int32) ([]EntryResponse, error) {
	rows, err := s.repo.search(ctx, agentID, from, to, q, limit)
	if err != nil {
		return nil, err
	}
	out := make([]EntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toEntryResponse(r.AgentID, r.Source, r.Level, r.Message, r.Count, r.FirstSeen, r.LastSeen, r.ReceivedAt)
	}
	return out, nil
}

func toEntryResponse(agentID uuid.UUID, source, level, message string, count int32, firstSeen, lastSeen, receivedAt time.Time) EntryResponse {
	return EntryResponse{
		AgentID:    agentID.String(),
		Source:     source,
		Level:      level,
		Message:    message,
		Count:      count,
		FirstSeen:  firstSeen,
		LastSeen:   lastSeen,
		ReceivedAt: receivedAt,
	}
}

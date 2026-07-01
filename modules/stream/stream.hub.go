package stream

import (
	"sync"
	"time"
)

type MetricEvent struct {
	AgentID    string            `json:"agent_id"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	RecordedAt time.Time         `json:"recorded_at"`
}

type LogEvent struct {
	AgentID  string    `json:"agent_id"`
	Source   string    `json:"source"`
	Level    string    `json:"level"`
	Message  string    `json:"message"`
	LastSeen time.Time `json:"last_seen"`
}

// Hub is an in-process pub/sub broker. Agents publish via the ingest path;
// SSE clients subscribe per agent ID. Slow consumers are dropped, never blocked.
type Hub struct {
	mu         sync.RWMutex
	metricSubs map[string][]chan MetricEvent
	logSubs    map[string][]chan LogEvent
}

func NewHub() *Hub {
	return &Hub{
		metricSubs: make(map[string][]chan MetricEvent),
		logSubs:    make(map[string][]chan LogEvent),
	}
}

func (h *Hub) PublishMetric(agentID, metricName string, value float64, labels map[string]string, recordedAt time.Time) {
	e := MetricEvent{AgentID: agentID, MetricName: metricName, Value: value, Labels: labels, RecordedAt: recordedAt}
	h.mu.RLock()
	subs := h.metricSubs[agentID]
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (h *Hub) PublishLog(agentID, source, level, message string, lastSeen time.Time) {
	e := LogEvent{AgentID: agentID, Source: source, Level: level, Message: message, LastSeen: lastSeen}
	h.mu.RLock()
	subs := h.logSubs[agentID]
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (h *Hub) subscribeMetrics(agentID string) (<-chan MetricEvent, func()) {
	ch := make(chan MetricEvent, 64)
	h.mu.Lock()
	h.metricSubs[agentID] = append(h.metricSubs[agentID], ch)
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		subs := h.metricSubs[agentID]
		for i, s := range subs {
			if s == ch {
				h.metricSubs[agentID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		h.mu.Unlock()
		close(ch)
	}
}

func (h *Hub) subscribeLogs(agentID string) (<-chan LogEvent, func()) {
	ch := make(chan LogEvent, 64)
	h.mu.Lock()
	h.logSubs[agentID] = append(h.logSubs[agentID], ch)
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		subs := h.logSubs[agentID]
		for i, s := range subs {
			if s == ch {
				h.logSubs[agentID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		h.mu.Unlock()
		close(ch)
	}
}

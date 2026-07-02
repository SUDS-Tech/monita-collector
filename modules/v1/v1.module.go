package v1

import (
	"github.com/bastion-framework/bast"

	"github.com/SUDS-Tech/monita-collector/modules/logs"
	"github.com/SUDS-Tech/monita-collector/modules/metrics"
)

func New(
	metricsSvc *metrics.Service,
	logsSvc *logs.Service,
	agentsSvc fingerprintWriter,
	rotator tokenRotator,
	agentGuard bast.Guard,
) bast.Module {
	c := newController(metricsSvc, logsSvc, agentsSvc, rotator, agentGuard)
	return bast.Module{
		Prefix:     "/v1",
		Controller: c,
		Doc: bast.ModuleDoc{
			Name:        "v1",
			Description: "Agent wire protocol v1 — metrics, logs, fingerprint, token rotation.",
		},
	}
}

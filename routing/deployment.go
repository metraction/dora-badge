package routing

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Tiktai/dora-badge/integrations"
	"github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"

	"github.com/Tiktai/dora-badge/logic"
	"github.com/Tiktai/dora-badge/model"
)

type HttpHandler struct {
	config *model.Config
}

func NewHttpHandler(config *model.Config) *HttpHandler {
	return &HttpHandler{config: config}
}

func (h *HttpHandler) HandleDeploymentsFrequency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	// Calculate the first and last day of the current month
	now := time.Now()
	// Extract project from URL path: /df/{project}
	project := r.PathValue("project")
	if project == "" {
		http.Error(w, "project not specified in path (use /v1/{project})/df", http.StatusBadRequest)
		return
	}
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := startDate.AddDate(0, 1, 0)
	endDate := nextMonth.Add(-time.Second) // Last second of the current month

	errors := make(chan any)

	// Query deployments per month
	devlake := integrations.NewDevlakeIntegration(h.config.DevLakeDSN)
	q := integrations.QueryDeployment{
		Project:     project,
		StartDate:   startDate.Format("2006-01-02"),
		FinishMonth: endDate.Format("2006-01-02"),
	}
	mapper := extension.NewChanSource(integrations.NewMono(q)).
		Via(flow.NewMap(devlake.QueryDeploymentsPerMonth, 1)).
		Via(flow.NewFilter(integrations.ErrorCollector[integrations.DeploymentMetric](errors), 1)).
		Via(flow.NewMap(integrations.Right[integrations.DeploymentMetric], 1))

	select {
	case response := <-mapper.Out():
		metrics := response.([]integrations.DeploymentMetric)
		deploymentCount := metrics[len(metrics)-1].DeploymentCount
		badge := logic.BadgeSVG("Deployments Frequency", fmt.Sprintf("%d", deploymentCount), "")
		io.WriteString(w, badge)

	case err := <-errors:
		badge := logic.BadgeSVG("Deployments Frequency", fmt.Sprintf("Error: %v", err), logic.BadgeErrorColor)
		io.WriteString(w, badge)
	}
}

package routing

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/metraction/dora-badge/integrations"
	"github.com/metraction/dora-badge/logic"
	"github.com/metraction/dora-badge/model"
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
	df, err := integrations.QueryDeploymentsPerMonth(
		h.config.DevLakeDSN,
		project,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
	if err != nil {
		log.Fatalf("QueryDeploymentsPerMonth failed: %v", err)
	}
	deploymentCount := df[len(df)-1].DeploymentCount
	badge := logic.BadgeSVG("Deployments Frequency", fmt.Sprintf("%d", deploymentCount), "")
	io.WriteString(w, badge)
}

package routing

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Tiktai/dora-badge/integrations"
	"github.com/Tiktai/dora-badge/logic"
)

// HandleLeadTimeForChanges serves a badge for lead time for changes for the current month for a project.
func (h *HttpHandler) HandleLeadTimeForChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	// Calculate the first and last day of the current month
	now := time.Now()
	project := ""
	if len(r.URL.Path) > len("/ltc/") {
		project = r.URL.Path[len("/ltc/"):]
	}
	if project == "" {
		http.Error(w, "project not specified in path (use /ltc/{project})", http.StatusBadRequest)
		return
	}
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := startDate.AddDate(0, 1, 0)
	endDate := nextMonth.Add(-time.Second)
	leadTime, err := integrations.QueryLeadTimeForChanges(
		h.config.DevLakeDSN,
		project,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		"2023",
	)
	if err != nil {
		log.Fatalf("QueryLeadTimeForChanges failed: %v", err)
	}
	badgeColor := ""
	if leadTime == "N/A. Please check if you have collected deployments/pull_requests." {
		badgeColor = logic.BadgeWarningColor
	}
	badge := logic.BadgeSVG("lead_time_for_changes", leadTime, badgeColor)
	io.WriteString(w, badge)
}

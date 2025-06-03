package routing

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/metraction/dora-badge/integrations"
	"github.com/metraction/dora-badge/logic"
)

// HandleLeadTimeForChanges serves a badge for lead time for changes for the current month for a project.
func (h *HttpHandler) HandleLeadTimeForChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	// Calculate the first and last day of the current month
	project := r.PathValue("project")
	if project == "" {
		http.Error(w, "project not specified in path (use /ltfc/{project})", http.StatusBadRequest)
		return
	}
	startDate, endDate := thisMonth()
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
	badge := logic.BadgeSVG("Lead Time for Changes", leadTime, badgeColor)
	io.WriteString(w, badge)
}

/*
HandleLeadTimeForChangesStats returns list of LeadTimeForChangesStats for a project in a given period.
*/
func (h *HttpHandler) HandleLeadTimeForChangesStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	project := r.PathValue("project")
	if project == "" {
		http.Error(w, "project not specified in path (use /ltfc-stats/{project})", http.StatusBadRequest)
		return
	}
	startDate, endDate := thisMonth()
	stats, err := integrations.QueryLeadTimeForChangesStats(
		h.config.DevLakeDSN,
		project,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		"2023",
	)
	if err != nil {
		log.Fatalf("QueryLeadTimeForChangesStats failed: %v", err)
	}
	for _, stat := range stats {
		d := time.Duration(stat.PrCycleTime) * time.Minute
		io.WriteString(w, fmt.Sprintf("PR sha: %s | Merged: %s | Finished: %s | Cycle: %s\n",
			stat.MergeCommitSHA[:7], stat.MergedDate, stat.FinishedDate, d))
	}
}

func thisMonth() (time.Time, time.Time) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := startDate.AddDate(0, 1, 0)
	endDate := nextMonth.Add(-time.Second)
	return startDate, endDate
}

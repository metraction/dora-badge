package integrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
)

// TestFetchDevLakeProjects tests FetchDevLakeProjects for correct project list retrieval.
func TestFetchDevLakeProjects(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping integration test")
	}

	projects, err := FetchDevLakeProjects(dsn)
	if err != nil {
		t.Fatalf("FetchDevLakeProjects failed: %v", err)
	}

	if len(projects) == 0 {
		t.Error("Expected at least one project, got 0")
	}
	for _, p := range projects {
		if p == "" {
			t.Error("Project name should not be empty")
		}
		log.Println(p)
	}
}

// Integration test for QueryLeadTimeForChanges
func TestQueryLeadTimeForChanges(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping integration test")
	}

	project := "Tiktai-handler" // Adjust as needed for your test DB
	startDate := "2025-05-01"
	finishMonth := "2025-05-31"
	doraReport := "2023"

	result, err := QueryLeadTimeForChanges(dsn, project, startDate, finishMonth, doraReport)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("QueryLeadTimeForChanges failed: %v", err)
	}
	if result == "" {
		t.Error("Expected a non-empty lead time result string")
	}
	fmt.Printf("Lead Time for Changes: %s\n", result)
}

// Integration test for QueryDeploymentsPerMonth
func TestQueryDeploymentsPerMonth(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN") // e.g. "user:pass@tcp(127.0.0.1:3306)/lake"
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping integration test")
	}

	project := "Tiktai"       // Change as appropriate or make dynamic
	startDate := "2025-05-01" // No time filter for test

	finishMonth := "2025-05-31" // or any month-end date you want to test inclusivity
	metrics, err := QueryDeploymentsPerMonth(dsn, project, startDate, finishMonth)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("QueryDeploymentsPerMonth failed: %v", err)
	}

	for _, m := range metrics {
		fmt.Printf("Month: %s, Deployments: %d\n", m.Month, m.DeploymentCount)
	}
}

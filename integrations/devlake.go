package integrations

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// DeploymentMetric holds the result for deployments per month
type DeploymentMetric struct {
	Month           string
	DeploymentCount int
}

const _pr_stats = `
    SELECT 
        distinct pr.id,
        pr.merge_commit_sha,
        pr.merged_date,
        cdc.finished_date,
        ppm.pr_cycle_time
    FROM
        pull_requests pr 
        join project_pr_metrics ppm on ppm.id = pr.id
        join project_mapping pm on pr.base_repo_id = pm.row_id and pm.` + "`table`" + ` = 'repos'
        join cicd_deployment_commits cdc on ppm.deployment_commit_id = cdc.id
    WHERE
      pm.project_name in (?) 
        and pr.merged_date is not null
        and ppm.pr_cycle_time is not null
        and cdc.finished_date >= ?
        and cdc.finished_date <= ?
`

// FetchDevLakeProjects returns a list of unique project names from the project_mapping table.
func FetchDevLakeProjects(dsn string) ([]string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT project_name FROM project_mapping ORDER BY project_name")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan project name: %w", err)
		}
		projects = append(projects, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

// convertDSNIfNeeded checks if the DSN is in URI format and converts it to Go MySQL driver format if needed.
func convertDSNIfNeeded(dsn string) string {
	if strings.HasPrefix(dsn, "mysql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return dsn // fallback to original if parsing fails
		}
		user := ""
		if u.User != nil {
			user = u.User.Username()
			if pass, ok := u.User.Password(); ok {
				user += ":" + pass
			}
		}
		host := u.Host
		// Remove possible port if present
		if !strings.Contains(host, ":") {
			host += ":3306"
		}
		// Remove leading slash from path to get dbname
		dbname := strings.TrimPrefix(u.Path, "/")
		params := u.RawQuery
		res := user + "@tcp(" + host + ")/" + dbname
		if params != "" {
			res += "?" + params
		}
		return res
	}
	return dsn
}

// QueryDeploymentsPerMonth connects to MySQL and executes the deployments per month metric query.
// startDate and finishMonth should be in 'YYYY-MM-DD' format, e.g. '2024-01-01'.
// Only months between startDate and finishMonth (inclusive) are returned.
// QueryLeadTimeForChanges calculates the median lead time for changes for a project in a given period and returns the result string.
// doraReport should be '2023' or '2021' to select the thresholds.
func QueryLeadTimeForChanges(dsn string, project string, startDate string, finishMonth string, doraReport string) (string, error) {
	log.Println("QueryLeadTimeForChanges", project, startDate, finishMonth, doraReport)
	dsn = convertDSNIfNeeded(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", fmt.Errorf("failed to connect to db: %w", err)
	}
	defer db.Close()

	query := `
with _pr_stats as (
    ` + _pr_stats + `
),

_median_change_lead_time_ranks as(
    SELECT *, percent_rank() over(order by pr_cycle_time) as ranks
    FROM _pr_stats
),

_median_change_lead_time as(
    SELECT max(pr_cycle_time) as median_change_lead_time
    FROM _median_change_lead_time_ranks
    WHERE ranks <= 0.5
)

SELECT 
  CASE
    WHEN (? = '2023') THEN
            CASE
                WHEN median_change_lead_time < 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(elite)')
                WHEN median_change_lead_time < 7 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(high)')
                WHEN median_change_lead_time < 30 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(medium)')
                WHEN median_change_lead_time >= 30 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(low)')
                ELSE 'N/A. Please check if you have collected deployments/pull_requests.'
                END
    WHEN (? = '2021') THEN
          CASE
                WHEN median_change_lead_time < 60 THEN CONCAT(round(median_change_lead_time/60,1), '(elite)')
                WHEN median_change_lead_time < 7 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(high)')
                WHEN median_change_lead_time < 180 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(medium)')
                WHEN median_change_lead_time >= 180 * 24 * 60 THEN CONCAT(round(median_change_lead_time/60,1), '(low)')
                ELSE 'N/A. Please check if you have collected deployments/pull_requests.'
                END
    ELSE 'N/A. Please check if you have collected deployments/pull_requests.'
  END as lead_time_for_changes
FROM _median_change_lead_time`

	row := db.QueryRow(query, project, startDate, finishMonth, doraReport, doraReport)
	var result string
	if err := row.Scan(&result); err != nil {
		return "", fmt.Errorf("failed to scan lead time for changes: %w", err)
	}
	return result, nil
}

type LeadTimeForChangesStats struct {
	PrId           string
	MergeCommitSHA string
	MergedDate     string
	FinishedDate   string
	PrCycleTime    int
}

func QueryLeadTimeForChangesStats(dsn string, project string, startDate string, finishMonth string, doraReport string) ([]LeadTimeForChangesStats, error) {
	dsn = convertDSNIfNeeded(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	defer db.Close()

	query := _pr_stats

	rows, err := db.Query(query, project, startDate, finishMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var stats []LeadTimeForChangesStats
	for rows.Next() {
		var stat LeadTimeForChangesStats
		if err := rows.Scan(
			&stat.PrId,
			&stat.MergeCommitSHA,
			&stat.MergedDate,
			&stat.FinishedDate,
			&stat.PrCycleTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func QueryDeploymentsPerMonth(dsn string, project string, startDate string, finishMonth string) ([]DeploymentMetric, error) {
	dsn = convertDSNIfNeeded(dsn)
	// Open connection to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	defer db.Close()

	// The metric query, using only parameter placeholders
	query := `
WITH _deployments AS (
    SELECT 
        date_format(deployment_finished_date,'%y/%m') as month,
        count(cicd_deployment_id) as deployment_count
    FROM (
        SELECT
            cdc.cicd_deployment_id,
            max(cdc.finished_date) as deployment_finished_date
        FROM cicd_deployment_commits cdc
        JOIN project_mapping pm on cdc.cicd_scope_id = pm.row_id and pm.table = 'cicd_scopes'
        WHERE
            pm.project_name = ?
            and cdc.result = 'SUCCESS'
            and cdc.environment = 'PRODUCTION'
            and cdc.finished_date >= ?
        GROUP BY 1
    ) _production_deployments
    GROUP BY 1
)
SELECT 
    cm.month, 
    CASE WHEN d.deployment_count IS NULL THEN 0 ELSE d.deployment_count END as deployment_count
FROM 
    calendar_months cm
    LEFT JOIN _deployments d on cm.month = d.month
WHERE cm.month_timestamp >= ?
  AND cm.month_timestamp <= ?`

	rows, err := db.Query(query, project, startDate, startDate, finishMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []DeploymentMetric
	for rows.Next() {
		var m DeploymentMetric
		if err := rows.Scan(&m.Month, &m.DeploymentCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

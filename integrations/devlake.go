package integrations

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// DeploymentMetric holds the result for deployments per month
type DeploymentMetric struct {
	Month           string
	DeploymentCount int
}

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

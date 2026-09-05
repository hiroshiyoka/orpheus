package storage

import (
	"database/sql"
	"time"
)

type Metric struct {
	ID            int64
	ProjectID     int64
	PeriodStart   time.Time
	RequestsCount int
	ErrorCount    int
	P50ResponseMS *int
	P99ResponseMS *int
}

func InsertMetric(db *sql.DB, metric Metric) (Metric, error) {
	result, err := db.Exec(`
		INSERT INTO metrics (project_id, period_start, requests_count, error_count, p50_response_ms, p99_response_ms)
		VALUES (?, ?, ?, ?, ?, ?)`,
		metric.ProjectID, metric.PeriodStart, metric.RequestsCount, metric.ErrorCount,
		metric.P50ResponseMS, metric.P99ResponseMS)
	if err != nil {
		return Metric{}, err
	}

	metric.ID, err = result.LastInsertId()
	if err != nil {
		return Metric{}, err
	}
	return getMetric(db, metric.ID)
}

func GetMetricsByProject(db *sql.DB, projectID int64, since time.Time) ([]Metric, error) {
	rows, err := db.Query(`
		SELECT id, project_id, period_start, requests_count, error_count, p50_response_ms, p99_response_ms
		FROM metrics
		WHERE project_id = ? AND period_start >= ?
		ORDER BY period_start, id`, projectID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		metric, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func getMetric(db *sql.DB, id int64) (Metric, error) {
	return scanMetric(db.QueryRow(`
		SELECT id, project_id, period_start, requests_count, error_count, p50_response_ms, p99_response_ms
		FROM metrics WHERE id = ?`, id))
}

func scanMetric(row rowScanner) (Metric, error) {
	var metric Metric
	var p50, p99 sql.NullInt64
	err := row.Scan(&metric.ID, &metric.ProjectID, &metric.PeriodStart, &metric.RequestsCount,
		&metric.ErrorCount, &p50, &p99)
	if err != nil {
		return Metric{}, err
	}
	if p50.Valid {
		value := int(p50.Int64)
		metric.P50ResponseMS = &value
	}
	if p99.Valid {
		value := int(p99.Int64)
		metric.P99ResponseMS = &value
	}
	return metric, nil
}

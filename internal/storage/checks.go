package storage

import (
	"database/sql"
	"time"
)

type Check struct {
	ID             int64
	ProjectID      int64
	CheckedAt      time.Time
	IsUp           bool
	StatusCode     *int
	ResponseTimeMS *int
	ErrorMessage   *string
}

func InsertCheck(db *sql.DB, check Check) (Check, error) {
	var result sql.Result
	var err error
	if check.CheckedAt.IsZero() {
		result, err = db.Exec(`
			INSERT INTO checks (project_id, is_up, status_code, response_time_ms, error_message)
			VALUES (?, ?, ?, ?, ?)`, check.ProjectID, check.IsUp, check.StatusCode,
			check.ResponseTimeMS, check.ErrorMessage)
	} else {
		result, err = db.Exec(`
			INSERT INTO checks (project_id, checked_at, is_up, status_code, response_time_ms, error_message)
			VALUES (?, ?, ?, ?, ?, ?)`, check.ProjectID, check.CheckedAt, check.IsUp,
			check.StatusCode, check.ResponseTimeMS, check.ErrorMessage)
	}
	if err != nil {
		return Check{}, err
	}

	check.ID, err = result.LastInsertId()
	if err != nil {
		return Check{}, err
	}
	return getCheck(db, check.ID)
}

func GetChecksByProject(db *sql.DB, projectID int64, since time.Time) ([]Check, error) {
	rows, err := db.Query(`
		SELECT id, project_id, checked_at, is_up, status_code, response_time_ms, error_message
		FROM checks
		WHERE project_id = ? AND checked_at >= ?
		ORDER BY checked_at, id`, projectID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []Check
	for rows.Next() {
		check, err := scanCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return checks, nil
}

func getCheck(db *sql.DB, id int64) (Check, error) {
	return scanCheck(db.QueryRow(`
		SELECT id, project_id, checked_at, is_up, status_code, response_time_ms, error_message
		FROM checks WHERE id = ?`, id))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCheck(row rowScanner) (Check, error) {
	var check Check
	var statusCode, responseTime sql.NullInt64
	var errorMessage sql.NullString
	err := row.Scan(&check.ID, &check.ProjectID, &check.CheckedAt, &check.IsUp,
		&statusCode, &responseTime, &errorMessage)
	if err != nil {
		return Check{}, err
	}
	if statusCode.Valid {
		value := int(statusCode.Int64)
		check.StatusCode = &value
	}
	if responseTime.Valid {
		value := int(responseTime.Int64)
		check.ResponseTimeMS = &value
	}
	if errorMessage.Valid {
		check.ErrorMessage = &errorMessage.String
	}
	return check, nil
}

package detector

import (
	"database/sql"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func ConsecutiveFailures(db *sql.DB, projectID int64, n int) (bool, error) {
	if n <= 0 {
		return false, nil
	}
	rows, err := db.Query(`SELECT is_up FROM checks WHERE project_id = ? ORDER BY checked_at DESC, id DESC LIMIT ?`, projectID, n)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var isUp bool
		if err := rows.Scan(&isUp); err != nil {
			return false, err
		}
		if isUp {
			return false, nil
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return count == n, nil
}

func CheckDowntime(db *sql.DB, projectID int64, n int) (*storage.Incident, error) {
	ok, err := ConsecutiveFailures(db, projectID, n)
	if err != nil || !ok {
		return nil, err
	}
	var id int64
	err = db.QueryRow(`SELECT id FROM incidents WHERE project_id = ? AND type = 'downtime' AND resolved_at IS NULL LIMIT 1`, projectID).Scan(&id)
	if err == nil {
		return nil, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	incident, err := storage.CreateIncident(db, storage.Incident{ProjectID: projectID, Type: "downtime", StartedAt: time.Now().UTC()})
	if err != nil {
		return nil, err
	}
	return &incident, nil
}

func ResolveDowntime(db *sql.DB, projectID int64, isUp bool) (*storage.Incident, error) {
	if !isUp {
		return nil, nil
	}
	var id int64
	var startedAt time.Time
	err := db.QueryRow(`SELECT id, started_at FROM incidents WHERE project_id = ? AND type = 'downtime' AND resolved_at IS NULL LIMIT 1`, projectID).Scan(&id, &startedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := storage.ResolveIncident(db, id, now); err != nil {
		return nil, err
	}
	return &storage.Incident{ID: id, ProjectID: projectID, Type: "downtime", StartedAt: startedAt, ResolvedAt: &now}, nil
}

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

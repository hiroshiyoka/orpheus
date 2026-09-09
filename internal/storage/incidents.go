package storage

import (
	"database/sql"
	"time"
)

type Incident struct {
	ID          int64
	ProjectID   int64
	Type        string
	StartedAt   time.Time
	ResolvedAt  *time.Time
	Description *string
}

func CreateIncident(db *sql.DB, incident Incident) (Incident, error) {
	result, err := db.Exec(`
		INSERT INTO incidents (project_id, type, started_at, resolved_at, description)
		VALUES (?, ?, ?, ?, ?)`,
		incident.ProjectID, incident.Type, incident.StartedAt,
		incident.ResolvedAt, incident.Description)
	if err != nil {
		return Incident{}, err
	}

	incident.ID, err = result.LastInsertId()
	if err != nil {
		return Incident{}, err
	}
	return getIncident(db, incident.ID)
}

func ResolveIncident(db *sql.DB, id int64, resolvedAt time.Time) error {
	result, err := db.Exec(`UPDATE incidents SET resolved_at = ? WHERE id = ?`, resolvedAt, id)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetActiveIncidents(db *sql.DB) ([]Incident, error) {
	rows, err := db.Query(`
		SELECT id, project_id, type, started_at, resolved_at, description
		FROM incidents
		WHERE resolved_at IS NULL
		ORDER BY started_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		incident, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, incident)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return incidents, nil
}

func ListIncidents(db *sql.DB, projectID int64) ([]Incident, error) {
	rows, err := db.Query(`
		SELECT id, project_id, type, started_at, resolved_at, description
		FROM incidents
		WHERE project_id = ?
		ORDER BY started_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		incident, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, incident)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return incidents, nil
}

func getIncident(db *sql.DB, id int64) (Incident, error) {
	return scanIncident(db.QueryRow(`
		SELECT id, project_id, type, started_at, resolved_at, description
		FROM incidents WHERE id = ?`, id))
}

func scanIncident(row rowScanner) (Incident, error) {
	var incident Incident
	var resolvedAt sql.NullTime
	var description sql.NullString
	err := row.Scan(&incident.ID, &incident.ProjectID, &incident.Type,
		&incident.StartedAt, &resolvedAt, &description)
	if err != nil {
		return Incident{}, err
	}
	if resolvedAt.Valid {
		incident.ResolvedAt = &resolvedAt.Time
	}
	if description.Valid {
		incident.Description = &description.String
	}
	return incident, nil
}

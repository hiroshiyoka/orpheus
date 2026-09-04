package storage

import (
	"database/sql"
	"time"
)

type Project struct {
	ID                   int64
	Name                 string
	URL                  string
	CloudflareZoneID     *string
	CheckIntervalSeconds int
	IsActive             bool
	CreatedAt            time.Time
}

func CreateProject(db *sql.DB, project Project) (Project, error) {
	result, err := db.Exec(`
		INSERT INTO projects (name, url, cloudflare_zone_id, check_interval_seconds, is_active)
		VALUES (?, ?, ?, ?, ?)`, project.Name, project.URL, project.CloudflareZoneID,
		project.CheckIntervalSeconds, project.IsActive)
	if err != nil {
		return Project{}, err
	}

	project.ID, err = result.LastInsertId()
	if err != nil {
		return Project{}, err
	}
	return GetProject(db, project.ID)
}

func GetProject(db *sql.DB, id int64) (Project, error) {
	var project Project
	var zoneID sql.NullString
	err := db.QueryRow(`
		SELECT id, name, url, cloudflare_zone_id, check_interval_seconds, is_active, created_at
		FROM projects WHERE id = ?`, id).Scan(
		&project.ID, &project.Name, &project.URL, &zoneID,
		&project.CheckIntervalSeconds, &project.IsActive, &project.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	if zoneID.Valid {
		project.CloudflareZoneID = &zoneID.String
	}
	return project, nil
}

func ListProjects(db *sql.DB) ([]Project, error) {
	rows, err := db.Query(`
		SELECT id, name, url, cloudflare_zone_id, check_interval_seconds, is_active, created_at
		FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var zoneID sql.NullString
		if err := rows.Scan(&project.ID, &project.Name, &project.URL, &zoneID,
			&project.CheckIntervalSeconds, &project.IsActive, &project.CreatedAt); err != nil {
			return nil, err
		}
		if zoneID.Valid {
			project.CloudflareZoneID = &zoneID.String
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

func UpdateProject(db *sql.DB, project Project) error {
	result, err := db.Exec(`
		UPDATE projects
		SET name = ?, url = ?, cloudflare_zone_id = ?, check_interval_seconds = ?, is_active = ?
		WHERE id = ?`, project.Name, project.URL, project.CloudflareZoneID,
		project.CheckIntervalSeconds, project.IsActive, project.ID)
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

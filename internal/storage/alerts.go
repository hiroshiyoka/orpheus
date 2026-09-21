package storage

import "database/sql"

func AlertSent(db *sql.DB, incidentID int64, channel, message string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM alerts_sent
			WHERE incident_id = ? AND channel = ? AND message = ?
		)`, incidentID, channel, message).Scan(&exists)
	return exists, err
}

func InsertAlert(db *sql.DB, incidentID int64, channel, message string) error {
	_, err := db.Exec(`
		INSERT INTO alerts_sent (incident_id, channel, message)
		VALUES (?, ?, ?)`, incidentID, channel, message)
	return err
}

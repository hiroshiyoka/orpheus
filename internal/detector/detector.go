package detector

import "database/sql"

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

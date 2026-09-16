package rule002

import "database/sql"

func query(db *sql.DB, ids []int) {
	for range ids {
		db.Query("SELECT id FROM users WHERE id = ?", 1) // want `SQL002: avoid database queries inside loops; batch or join instead`
	}

	db.Query("SELECT id FROM users WHERE id IN (?)", ids)
}

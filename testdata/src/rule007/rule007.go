package rule007

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT id FROM orders WHERE status = 'PAID'") // want `SQL007: parameterize literal 'PAID' in the SQL statement`
	db.Query("SELECT id FROM orders WHERE status = ?", "PAID")
	db.Query("SELECT id FROM orders WHERE status = $1", "PAID")
}

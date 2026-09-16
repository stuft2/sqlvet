package rule009

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT id, message FROM logs ORDER BY created_at DESC") // want `SQL009: bound SELECT results with LIMIT, FETCH, TOP, or pagination`
	db.Query("SELECT id, message FROM logs ORDER BY created_at DESC LIMIT 100")
	db.QueryRow("SELECT COUNT(*) FROM logs")
}

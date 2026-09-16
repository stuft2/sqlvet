package rule010

import "database/sql"

func explain(db *sql.DB) {
	db.Query("EXPLAIN SELECT id FROM users WHERE id = ?", 1) // want `SQL010: EXPLAIN result is discarded; inspect the returned execution plan`

	rows, _ := db.Query("EXPLAIN SELECT id FROM users WHERE id = ?", 1)
	_ = rows
}

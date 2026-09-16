package rule001

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT * FROM users") // want `SQL001: avoid SELECT \*; select only the columns you need`
	db.Query("SELECT id, name FROM users")
	db.Query("SELECT COUNT(*) FROM users")
}

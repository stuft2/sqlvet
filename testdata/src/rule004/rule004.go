package rule004

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT u.id, o.total FROM users u LEFT JOIN orders o ON o.user_id = u.id WHERE o.status = ?", "paid") // want `SQL004: LEFT JOIN alias o is null-rejected in WHERE; use INNER JOIN or move the predicate`

	db.Query(`SELECT u.id, o.total
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id
		WHERE o.id IS NULL`)
}

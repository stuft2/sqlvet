package rule006

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT id FROM orders WHERE DATE(order_date) = ?", "2026-09-15") // want `SQL006: avoid applying DATE to a column in WHERE; use a searchable range predicate`
	db.Query("SELECT id FROM orders WHERE order_date >= ? AND order_date < ?", "2026-09-15", "2026-09-16")
}

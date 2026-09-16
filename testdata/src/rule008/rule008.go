package rule008

import "database/sql"

func query(db *sql.DB) {
	db.Query("SELECT DISTINCT user_id FROM orders") // want `SQL008: review SELECT DISTINCT; fix duplicate joins or data at the source`
	db.Query("SELECT user_id, SUM(amount) FROM orders GROUP BY user_id")
}

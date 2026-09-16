package rule005

import "database/sql"

func updateWithoutTransaction(db *sql.DB) {
	db.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1)
	db.Exec("UPDATE transactions SET status = ? WHERE id = ?", "PAID", 10) // want `SQL005: multiple writes through \*sql.DB should be wrapped in a transaction`
}

func updateWithTransaction(db *sql.DB) {
	tx, _ := db.Begin()
	tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1)
	tx.Exec("UPDATE transactions SET status = ? WHERE id = ?", "PAID", 10)
	tx.Commit()
}

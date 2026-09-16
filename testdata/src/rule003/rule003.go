package rule003

import "database/sql"

func migrate(db *sql.DB) {
	db.Exec("CREATE TABLE orders (id INTEGER PRIMARY KEY, customer_id INTEGER, FOREIGN KEY (customer_id) REFERENCES customers(id))") // want `SQL003: foreign key customer_id has no supporting index in this DDL statement`

	db.Exec(`
		CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			order_id INTEGER,
			FOREIGN KEY (order_id) REFERENCES orders(id)
		);
		CREATE INDEX idx_items_order_id ON items(order_id);
	`)
}

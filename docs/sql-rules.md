# SQL Review Rules

Use these rules during SQL authoring and code review to improve correctness,
performance, and maintainability. Examples are illustrative; adapt parameter
syntax and database-specific features to the SQL dialect in use.

## 1. Select only the columns you need

Avoid `SELECT *`. Returning unnecessary columns increases I/O, memory use, and
coupling to the table schema.

```sql
SELECT id, user_id, order_date, amount
FROM orders
WHERE user_id = :user_id;
```

## 2. Avoid N+1 queries

Do not fetch a collection and then issue one query per row. Prefer a join, a
batched query, or an appropriate bulk-loading feature.

```sql
SELECT u.id, u.name, o.id AS order_id, o.amount
FROM users AS u
LEFT JOIN orders AS o ON o.user_id = u.id
WHERE u.id IN (:user_ids);
```

## 3. Add indexes that support query access patterns

Consider indexes for columns frequently used in `WHERE`, `JOIN`, and `ORDER BY`
clauses. Confirm that an index matches real query patterns; unnecessary indexes
consume storage and slow writes.

```sql
CREATE INDEX idx_orders_customer_id ON orders (customer_id);

SELECT id, customer_id, amount
FROM orders
WHERE customer_id = :customer_id;
```

## 4. Use the join type that matches the intended result

Use `INNER JOIN` when only matching rows should be returned. Use `LEFT JOIN`
when every row from the left side must be retained, including unmatched rows.

```sql
SELECT u.name, o.total
FROM users AS u
INNER JOIN orders AS o ON o.user_id = u.id;
```

## 5. Wrap related writes in a transaction

Operations that must succeed or fail together belong in one transaction. Roll
back the transaction when any operation fails.

```sql
BEGIN;

UPDATE accounts
SET balance = balance - :amount
WHERE id = :account_id;

UPDATE transactions
SET status = 'PAID'
WHERE id = :transaction_id;

COMMIT;
```

## 6. Keep indexed columns searchable in predicates

Avoid wrapping an indexed column in a function when an equivalent range
predicate is available. A function may prevent the database from using a normal
index on that column.

```sql
SELECT id, order_date, amount
FROM orders
WHERE order_date >= :day_start
  AND order_date < :next_day_start;
```

## 7. Parameterize values

Do not embed user input or frequently changing values directly in SQL. Bind
parameters improve safety, reuse, and maintainability.

```sql
SELECT id, status, created_at
FROM orders
WHERE status = :status
  AND created_at >= :created_after;
```

## 8. Do not use `DISTINCT` to hide incorrect joins or duplicate data

Before adding `DISTINCT`, identify why duplicate rows exist. Correct the join,
data model, or aggregation. Use `DISTINCT` only when unique result rows are the
actual requirement.

```sql
SELECT user_id, SUM(amount) AS total_amount
FROM orders
WHERE amount > 0
GROUP BY user_id;
```

## 9. Bound large result sets

Avoid queries that can return an unlimited number of rows. Use a sensible limit
and deterministic ordering, or use pagination for larger result sets.

```sql
SELECT id, created_at, message
FROM logs
ORDER BY created_at DESC, id DESC
LIMIT :page_size;
```

`LIMIT` syntax varies by database. Prefer keyset pagination over large offsets
when browsing deep into a frequently changing or very large result set.

## 10. Inspect the execution plan

Use the database's execution-plan command to verify access paths, join
strategies, row estimates, and sorting behavior before expensive queries reach
production.

```sql
EXPLAIN
SELECT id, customer_id, amount
FROM orders
WHERE customer_id = :customer_id;
```

Use execution options that run the query, such as `EXPLAIN ANALYZE`, carefully:
they can execute writes or expensive workloads depending on the database.

## Automated checks

The sqlvet analyzer maps these guidelines to rule IDs `SQL001` through
`SQL010`. Static analysis cannot prove schema state or application intent, so
some checks intentionally target narrower signals:

- `SQL003` checks foreign keys declared without an index in the same DDL string.
- `SQL004` checks a `LEFT JOIN` whose right-side alias is null-rejected in
  `WHERE`.
- `SQL005` checks multiple writes made through `*sql.DB` in one Go function.
- `SQL010` checks an `EXPLAIN` query whose returned rows are discarded.

All other rules inspect compile-time SQL strings passed to supported
`database/sql` methods. Runtime-generated queries and calls through custom
interfaces are not analyzed.

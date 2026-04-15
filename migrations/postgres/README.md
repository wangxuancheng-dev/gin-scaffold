PostgreSQL migrations

- Put PostgreSQL-specific `*.up.sql` and `*.down.sql` files in this directory.
- Naming convention should follow timestamp order, e.g.:
  - `202501011200_create_users.up.sql`
  - `202501011200_create_users.down.sql`
- `cmd/migrate` auto-selects this directory when `--driver postgres`.


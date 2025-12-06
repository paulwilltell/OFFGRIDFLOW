# Database Migrations

This directory contains database migration files for OffGridFlow.

## Prerequisites

Install golang-migrate:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Usage

### Create a New Migration
```bash
# Linux/Mac
./scripts/migrate.sh create init_schema

# Windows
.\scripts\migrate.ps1 create init_schema
```

This creates two files:
- `000001_init_schema.up.sql` - Applied when migrating up
- `000001_init_schema.down.sql` - Applied when rolling back

### Apply Migrations
```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Apply only 1 migration
./scripts/migrate.sh up 1
```

### Rollback Migrations
```bash
# Rollback all migrations
./scripts/migrate.sh down

# Rollback only 1 migration
./scripts/migrate.sh down 1
```

### Check Current Version
```bash
./scripts/migrate.sh version
```

## Migration Files

Migrations are numbered sequentially:
- `000001_init_schema.up.sql`
- `000001_init_schema.down.sql`
- `000002_add_users.up.sql`
- `000002_add_users.down.sql`
- etc.

## Best Practices

1. **Always write DOWN migrations** - Every UP must have a corresponding DOWN
2. **Test migrations** - Test both UP and DOWN before committing
3. **Keep migrations small** - One logical change per migration
4. **Never edit applied migrations** - Create a new migration instead
5. **Use transactions** - Wrap DDL in `BEGIN;` / `COMMIT;`

## Example Migration

**000001_init_schema.up.sql**:
```sql
BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

COMMIT;
```

**000001_init_schema.down.sql**:
```sql
BEGIN;

DROP TABLE IF EXISTS users CASCADE;

COMMIT;
```

## Auto-Migration on Startup

The application automatically runs migrations on startup.

See: `cmd/api/main.go` for implementation.

## Troubleshooting

### Migration is stuck
```bash
# Check current version
./scripts/migrate.sh version

# Force to specific version (DANGEROUS - only use if you know what you're doing)
./scripts/migrate.sh force 5
```

### Reset database completely
```bash
# WARNING: This drops ALL data
./scripts/migrate.sh drop

# Then re-apply all migrations
./scripts/migrate.sh up
```

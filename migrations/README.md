# Database Migrations

This directory contains SQL migration scripts for the anti-bully database.

## How to Apply Migrations

### Method 1: Using psql (Recommended)

```bash
# From the backend directory
psql $DATABASE_URL -f migrations/001_add_reporter_fields.sql
```

### Method 2: Direct psql connection

```bash
# Connect to your database
psql -h <host> -U <username> -d <database> -p <port>

# Then run the migration
\i migrations/001_add_reporter_fields.sql
```

### Method 3: Using Go's AutoMigrate (Automatic)

The GORM AutoMigrate should handle this automatically when you start the server, unless:
- `SKIP_MIGRATE=1` is set in your environment
- There's a version mismatch between GORM and your PostgreSQL driver

If AutoMigrate is failing, use Method 1 or 2 instead.

## Migration List

- `001_add_reporter_fields.sql` - Adds `reporter_id` and `reporter_name` columns to `reports` table

## Troubleshooting

If you see the error:
```
ERROR: column "reporter_id" of relation "reports" does not exist
```

This means the migration hasn't been applied yet. Run the SQL migration manually using Method 1.

## Verifying Migration Success

After applying the migration, verify the columns exist:

```sql
\d reports
-- or
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'reports';
```

You should see `reporter_id` and `reporter_name` in the output.

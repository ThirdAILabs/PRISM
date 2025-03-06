# Versioned Migrations in Prism

Gorm comes with various migration utiliites, the only thing that's missing is the versioning. We're using a lightweight library called gormigrate which adds versioning on top of Gorm. The example here shows how it is used: https://github.com/go-gormigrate/gormigrate

## Running Migrations on Startup
Everytime the backend starts, it will check the current latest schema version against the database and perform any necessary migrations. Note: this is done by the backend, not the worker. Make sure to start the backend first to ensure all migrations are applied. 

## Running Migrations Manually
To run a migration manually run `go run ./prism/cmd/migrations/main.go --version <target schema version> --rollback`. The `--version` arg indicates which version to migrate to, the `--rollback` arg is only needed if the migration is a rollback.

## Adding New Versions
1. Add a new migration to the list of `gormigrate.Migration` structs in the migrator object in `migrations.go`.
2. In the struct you must define the ID, as well as the `Migrate` function. If the migration is reversible define the `Rollback` function as well.
3. For organization we will define these functions in `versions/migration_<version>.go` so that it is easy to find the code for the appropriate migration.
4. Never modify the code for a previous migration, unless it's to fix a bug. 
5. After adding the migration please test out both the `Migration` and `Rollback` functionality to ensure it is working correctly.

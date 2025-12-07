package migrations

import (
	"database/sql"
	"fmt"
	"log/slog"

	sqltool "github.com/cheatsnake/airstation/internal/pkg/sql"
)

func RunMigrations(db *sql.DB, log *slog.Logger) error {
	migrationTableExists, err := sqltool.TableExists(db, "migrations")
	if err != nil {
		return err
	}

	var currentVersion int

	if migrationTableExists {
		err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&currentVersion)
		if err != nil {
			return fmt.Errorf("failed to get current version: %w", err)
		}
	} else {
		currentVersion = 0
		log.Info("Fresh database, starting migrations from the beginning")
	}

	for _, migration := range migrations {
		if migration.Version > currentVersion {
			log.Info("Applying migration", "version", migration.Version, "name", migration.Name)

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction for migration %d: %w",
					migration.Version, err)
			}

			defer func() {
				if tx != nil {
					tx.Rollback()
				}
			}()

			if err := migration.Up(tx); err != nil {
				return fmt.Errorf("migration %d (%s) failed: %w",
					migration.Version, migration.Name, err)
			}

			if migrationTableExists || migration.Version >= 1 {
				_, err = tx.Exec(
					"INSERT INTO migrations (version, name) VALUES (?, ?)",
					migration.Version, migration.Name,
				)
				if err != nil {
					return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
				}
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
			}

			tx = nil // Remove defer rollback after successful commit

			log.Info("Migration applied successfully",
				"version", migration.Version,
				"name", migration.Name)
		}
	}

	var finalVersion int

	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&finalVersion)
	if err != nil {
		return fmt.Errorf("failed to get final version: %w", err)
	}

	if finalVersion > currentVersion {
		log.Info("Database migration completed",
			"from_version", currentVersion,
			"to_version", finalVersion,
			"migrations_applied", finalVersion-currentVersion)
	} else {
		log.Info("Database is up to date", "version", finalVersion)
	}

	return nil
}

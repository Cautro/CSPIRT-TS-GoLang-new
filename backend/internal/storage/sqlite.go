package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const sqliteDriverName = "sqlite"

func openSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open(sqliteDriverName, path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := configureSQLite(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA temp_store = MEMORY`,
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}

package storage

import "database/sql"

func (s *Storage) initSchema() error {
	if err := s.initUserStorage(); err != nil {
		return err
	}
	if err := s.initClassStorage(); err != nil {
		return err
	}
	if err := s.initNoteStorage(); err != nil {
		return err
	}
	if err := s.initComplaintStorage(); err != nil {
		return err
	}
	if err := s.initHTTPOnlyStorage(); err != nil {
		return err
	}
	if err := s.syncAllClasses(); err != nil {
		return err
	}
	if err := s.initEventsStorage(); err != nil {
		return err
	}
	if err := s.initBaseSchedulesStorage(); err != nil {
		return err
	}
	if err := s.initExceptionsSchedulesStorage(); err != nil {
		return err
	}
	if err := s.initPlannedSchedulesStorage(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) initBaseSchedulesStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS schedules (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		ClassID INTEGER NOT NULL,
		DayOfWeek TEXT NOT NULL,
		LessonNumber INTEGER NOT NULL,
		WeekType TEXT NOT NULL DEFAULT 'all',
		Subject TEXT NOT NULL,
		TeacherID INTEGER NOT NULL,
		Room INTEGER NOT NULL,
		StartTime TEXT NOT NULL,
		EndTime TEXT NOT NULL,
		Description TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (TeacherID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("schedules", "WeekType", "TEXT NOT NULL DEFAULT 'all'"); err != nil {
		return err
	}
	if err := s.ensureColumn("schedules", "Description", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_schedules_unique_lesson
		ON schedules(ClassID, DayOfWeek, LessonNumber, WeekType);
	`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_class_id ON schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_day ON schedules(DayOfWeek);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initExceptionsSchedulesStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS exceptions_schedules (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		ScheduleID INTEGER,
		ClassID INTEGER NOT NULL,
		Date TEXT NOT NULL,
		ChangeType TEXT NOT NULL CHECK (
			ChangeType IN (
				'cancel', 'replace', 'move', 'room_change', 
				'teacher_change', 'update', 'add', 'day_off', 
				'short_day', 'swap'
			)
		),
		NewSubject TEXT,
		Scope TEXT NOT NULL DEFAULT 'lesson',
		NewLessonNumber INTEGER,
		NewTeacherID INTEGER,
		NewRoom INTEGER,
		NewStartTime TEXT,
		NewEndTime TEXT,
		NewDescription TEXT,
		Reason TEXT NOT NULL DEFAULT '',
		CreatedAt TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (ScheduleID) REFERENCES schedules(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (NewTeacherID) REFERENCES users(Id) ON DELETE SET NULL
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("exceptions_schedules", "ClassID", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("exceptions_schedules", "NewDescription", "TEXT"); err != nil {
		return err
	}
	if err := s.ensureColumn("exceptions_schedules", "CreatedAt", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_exceptions_schedules_class_id ON exceptions_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_exceptions_schedules_date ON exceptions_schedules(Date);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_exceptions_schedules_schedule_id ON exceptions_schedules(ScheduleID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initPlannedSchedulesStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS planned_schedules (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		BaseScheduleID INTEGER,
		ClassID INTEGER NOT NULL,
		Date TEXT NOT NULL,
		LessonNumber INTEGER NOT NULL,
		Subject TEXT NOT NULL,
		ChangeType TEXT NOT NULL CHECK (
			ChangeType IN (
				'cancel', 'replace', 'move', 'room_change', 
				'teacher_change', 'update', 'add', 'day_off', 
				'short_day', 'swap'
			)
		),
		Scope TEXT NOT NULL DEFAULT 'lesson',
		TeacherID INTEGER NOT NULL,
		Room INTEGER NOT NULL,
		StartTime TEXT NOT NULL,
		EndTime TEXT NOT NULL,
		Description TEXT NOT NULL DEFAULT '',
		Reason TEXT NOT NULL DEFAULT '',
		CreatedAt TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (BaseScheduleID) REFERENCES schedules(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (TeacherID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("planned_schedules", "Scope", "TEXT NOT NULL DEFAULT 'lesson'"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_class_id ON planned_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_date ON planned_schedules(Date);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_base_id ON planned_schedules(BaseScheduleID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initUserStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS users (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT NOT NULL,
		FullName TEXT NOT NULL DEFAULT '[]',
		LastName TEXT NOT NULL,
		Login TEXT NOT NULL UNIQUE,
		Password TEXT NOT NULL,
		Rating INTEGER NOT NULL DEFAULT 500,
		Role TEXT NOT NULL,
		Class TEXT NOT NULL,
		ClassID INTEGER
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	if err := s.ensureColumn("users", "ClassID", "INTEGER"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_role_id ON users(ClassID, Role, Id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initEventsStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventsQuery := `
	CREATE TABLE IF NOT EXISTS events (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Status TEXT NOT NULL DEFAULT 'scheduled',
		RatingReward INTEGER NOT NULL DEFAULT 0,
		Description TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		StartedAt TEXT NOT NULL,
		Players TEXT NOT NULL,
		Classes TEXT NOT NULL DEFAULT '[]'
	);`

	if _, err := s.db.Exec(eventsQuery); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "RatingReward", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "Classes", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}

	eventPlayersQuery := `
	CREATE TABLE IF NOT EXISTS event_players (
		event_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		PRIMARY KEY (event_id, player_id),
		FOREIGN KEY (event_id) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(eventPlayersQuery); err != nil {
		return err
	}

	eventClassesQuery := `
	CREATE TABLE IF NOT EXISTS event_classes (
		event_id INTEGER NOT NULL,
		class_id INTEGER NOT NULL,
		PRIMARY KEY (event_id, class_id),
		FOREIGN KEY (event_id) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (class_id) REFERENCES classes(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(eventClassesQuery); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_event_players_player_id ON event_players(player_id);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_event_classes_class_id ON event_classes(class_id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initNoteStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS notes (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		TargetID INTEGER NOT NULL,
		AuthorID INTEGER NOT NULL,
		TargetName TEXT NOT NULL,
		AuthorName TEXT NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_target_id ON notes(TargetID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_author_id ON notes(AuthorID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initHTTPOnlyStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initComplaintStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS complaints (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		TargetID INTEGER NOT NULL,
		TargetName TEXT NOT NULL,
		AuthorID INTEGER NOT NULL,
		AuthorName TEXT NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		FOREIGN KEY (TargetID) REFERENCES users(Id) ON DELETE CASCADE,
		FOREIGN KEY (AuthorID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_target_id ON complaints(TargetID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_author_id ON complaints(AuthorID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) ensureColumn(table string, column string, definition string) error {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = s.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}

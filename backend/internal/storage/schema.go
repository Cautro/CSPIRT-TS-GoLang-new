package storage

import "database/sql"

func (s *Storage) initSchema() error {
	if err := s.initUserStorage(); err != nil {
		return err
	}
	if err := s.initClassStorage(); err != nil {
		return err
	}
	if err := s.initParallelsStorage(); err != nil {
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
	if err := s.initParamsForEventsStorage(); err != nil {
		return err
	}
	if err := s.initBaseSchedulesStorage(); err != nil {
		return err
	}
	if err := s.initCurrentSchedulesStorage(); err != nil {
		return err
	}
	if err := s.initPlannedSchedulesStorage(); err != nil {
		return err
	}
	if err := s.ensureCurrentSchedulesSeeded(); err != nil {
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

func (s *Storage) initCurrentSchedulesStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS current_schedules (
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
	if err := s.ensureColumn("current_schedules", "WeekType", "TEXT NOT NULL DEFAULT 'all'"); err != nil {
		return err
	}
	if err := s.ensureColumn("current_schedules", "Description", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_current_schedules_unique_lesson
		ON current_schedules(ClassID, DayOfWeek, LessonNumber, WeekType);
	`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_class_id ON current_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_teacher_id ON current_schedules(TeacherID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_day ON current_schedules(DayOfWeek);`); err != nil {
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
		Date TEXT NOT NULL DEFAULT '',
		DayOfWeek TEXT NOT NULL DEFAULT '',
		LessonNumber INTEGER NOT NULL,
		WeekType TEXT NOT NULL DEFAULT 'all',
		Subject TEXT NOT NULL,
		ChangeType TEXT NOT NULL DEFAULT 'update' CHECK (
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
	if err := s.ensureColumn("planned_schedules", "DayOfWeek", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("planned_schedules", "WeekType", "TEXT NOT NULL DEFAULT 'all'"); err != nil {
		return err
	}
	if err := s.ensureColumn("planned_schedules", "Scope", "TEXT NOT NULL DEFAULT 'lesson'"); err != nil {
		return err
	}
	if err := s.ensureColumn("planned_schedules", "Date", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("planned_schedules", "ChangeType", "TEXT NOT NULL DEFAULT 'update'"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_class_id ON planned_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_teacher_id ON planned_schedules(TeacherID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_day ON planned_schedules(DayOfWeek);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_base_id ON planned_schedules(BaseScheduleID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
		UPDATE planned_schedules
		SET DayOfWeek = COALESCE((SELECT DayOfWeek FROM schedules WHERE schedules.Id = planned_schedules.BaseScheduleID), DayOfWeek),
			WeekType = COALESCE((SELECT WeekType FROM schedules WHERE schedules.Id = planned_schedules.BaseScheduleID), WeekType)
		WHERE TRIM(DayOfWeek) = ''
			AND BaseScheduleID IS NOT NULL;
	`); err != nil {
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

func (s *Storage) initParamsForEventsStorage() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    query := `
    CREATE TABLE IF NOT EXISTS event_params (
        Id INTEGER PRIMARY KEY AUTOINCREMENT,
        EventID INTEGER NOT NULL,
		ClassID INTEGER NOT NULL,
		ExtraRatingReward INTEGER NOT NULL DEFAULT 0,
		Reason TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (EventID) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE
	);`

    if _, err := s.db.Exec(query); err != nil {
        return err
	}

	if err := s.ensureColumn(
		"event_params",
		"ExtraRatingReward",
		"INTEGER NOT NULL DEFAULT 0",
	); err != nil {
		return err
	}

	if err := s.ensureColumn(
		"event_params",
		"Reason",
		"TEXT NOT NULL DEFAULT ''",
	); err != nil {
		return err
	}

	if err := s.ensureColumn(
		"event_params",
		"ClassID",
		"INTEGER NOT NULL DEFAULT 0",
	); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_event_params_event_id
		ON event_params(EventID);
	`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_event_params_class_id
		ON event_params(ClassID);
	`); err != nil {
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

func (s *Storage) initParallelsStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// ВРЕМЕННО: Удаляем старую сломанную таблицу, чтобы SQLite создал её заново
	// (Если приложение запустится успешно, эту строку можно будет удалить)
	// s.db.Exec(`DROP TABLE IF EXISTS parallels;`)

	query := `
	CREATE TABLE IF NOT EXISTS parallels (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT NOT NULL,
		Rating INTEGER NOT NULL DEFAULT 0,
		BestClassID INTEGER NOT NULL,
		ClassID INTEGER,
		ParallelClassID INTEGER, -- ИСПРАВЛЕНО: Теперь колонка существует
		ValidClasses INTEGER NOT NULL,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (ParallelClassID) REFERENCES classes(Id) ON DELETE CASCADE
	);`
	
	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_parallels_class_pair ON parallels(ClassID, ParallelClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_parallels_class_id ON parallels(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_parallels_parallel_class_id ON parallels(ParallelClassID);`); err != nil {
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
package storage

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
	if err := syncAllClasses(s.DB); err != nil {
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

	query := `
	CREATE TABLE IF NOT EXISTS schedules (
		Id BIGSERIAL PRIMARY KEY,
		ClassID INTEGER NOT NULL,
		DayOfWeek TEXT NOT NULL,
		LessonNumber INTEGER NOT NULL,
		WeekType TEXT NOT NULL DEFAULT 'all',
		Subject TEXT NOT NULL,
		TeacherID INTEGER NOT NULL,
		Room INTEGER NOT NULL,
		StartTime TIMESTAMPTZ NOT NULL,
		EndTime TIMESTAMPTZ NOT NULL,
		Description TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (TeacherID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("schedules", "WeekType", "TEXT NOT NULL DEFAULT 'all'"); err != nil {
		return err
	}
	if err := s.ensureColumn("schedules", "Description", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_schedules_unique_lesson
		ON schedules(ClassID, DayOfWeek, LessonNumber, WeekType);
	`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_class_id ON schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_day ON schedules(DayOfWeek);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initCurrentSchedulesStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS current_schedules (
		Id BIGSERIAL PRIMARY KEY,
		ClassID INTEGER NOT NULL,
		DayOfWeek TEXT NOT NULL,
		LessonNumber INTEGER NOT NULL,
		WeekType TEXT NOT NULL DEFAULT 'all',
		Subject TEXT NOT NULL,
		TeacherID INTEGER NOT NULL,
		Room INTEGER NOT NULL,
		StartTime TIMESTAMPTZ NOT NULL,
		EndTime TIMESTAMPTZ NOT NULL,
		Description TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (TeacherID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	if err := s.ensureColumn("current_schedules", "WeekType", "TEXT NOT NULL DEFAULT 'all'"); err != nil {
		return err
	}
	if err := s.ensureColumn("current_schedules", "Description", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_current_schedules_unique_lesson
		ON current_schedules(ClassID, DayOfWeek, LessonNumber, WeekType);
	`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_class_id ON current_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_teacher_id ON current_schedules(TeacherID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_current_schedules_day ON current_schedules(DayOfWeek);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initPlannedSchedulesStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS planned_schedules (
		Id BIGSERIAL PRIMARY KEY,
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
		StartTime TIMESTAMPTZ NOT NULL,
		EndTime TIMESTAMPTZ NOT NULL,
		Description TEXT NOT NULL DEFAULT '',
		Reason TEXT NOT NULL DEFAULT '',
		CreatedAt TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (BaseScheduleID) REFERENCES schedules(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE,
		FOREIGN KEY (TeacherID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
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
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_class_id ON planned_schedules(ClassID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_teacher_id ON planned_schedules(TeacherID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_day ON planned_schedules(DayOfWeek);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_planned_schedules_base_id ON planned_schedules(BaseScheduleID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`
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

	query := `
	CREATE TABLE IF NOT EXISTS users (
		Id BIGSERIAL PRIMARY KEY,
		Avatar TEXT,
		Name TEXT NOT NULL,
		FullName JSONB NOT NULL DEFAULT '[]',
		LastName TEXT NOT NULL,
		Login TEXT NOT NULL UNIQUE,
		Password TEXT NOT NULL,
		Rating INTEGER NOT NULL DEFAULT 500,
		Role TEXT NOT NULL,
		Class TEXT NOT NULL,
		ClassID INTEGER
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}

	if err := s.ensureColumn("users", "ClassID", "INTEGER"); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(ClassID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_role_id ON users(ClassID, Role, Id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initEventsStorage() error {

	eventsQuery := `
	CREATE TABLE IF NOT EXISTS events (
		Id BIGSERIAL PRIMARY KEY,
		Title TEXT NOT NULL,
		Status TEXT NOT NULL DEFAULT 'scheduled',
		RatingReward INTEGER NOT NULL DEFAULT 0,
		Description TEXT NOT NULL,
		CreatedAt TIMESTAMPTZ NOT NULL,
		StartedAt TIMESTAMPTZ NOT NULL,
		Players TEXT NOT NULL,
		Classes JSONB NOT NULL DEFAULT '[]'
	);`

	if _, err := s.DB.Exec(eventsQuery); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "RatingReward", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "Classes", "JSONB NOT NULL DEFAULT '[]'"); err != nil {
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

	if _, err := s.DB.Exec(eventPlayersQuery); err != nil {
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

	if _, err := s.DB.Exec(eventClassesQuery); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_event_players_player_id ON event_players(player_id);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_event_classes_class_id ON event_classes(class_id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initParamsForEventsStorage() error {

	query := `
    CREATE TABLE IF NOT EXISTS event_params (
        Id BIGSERIAL PRIMARY KEY,
        EventID INTEGER NOT NULL,
		ClassID INTEGER NOT NULL,
		ExtraRatingReward INTEGER NOT NULL DEFAULT 0,
		Reason TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (EventID) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
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

	if _, err := s.DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_event_params_event_id
		ON event_params(EventID);
	`); err != nil {
		return err
	}

	if _, err := s.DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_event_params_class_id
		ON event_params(ClassID);
	`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initNoteStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS notes (
		Id BIGSERIAL PRIMARY KEY,
		TargetID INTEGER NOT NULL,
		AuthorID INTEGER NOT NULL,
		TargetName TEXT NOT NULL,
		AuthorName TEXT NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_target_id ON notes(TargetID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_author_id ON notes(AuthorID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initHTTPOnlyStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id BIGSERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initComplaintStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS complaints (
		Id BIGSERIAL PRIMARY KEY,
		TargetID INTEGER NOT NULL,
		TargetName TEXT NOT NULL,
		AuthorID INTEGER NOT NULL,
		AuthorName TEXT NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TIMESTAMPTZ NOT NULL,
		FOREIGN KEY (TargetID) REFERENCES users(Id) ON DELETE CASCADE,
		FOREIGN KEY (AuthorID) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_target_id ON complaints(TargetID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_author_id ON complaints(AuthorID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initParallelsStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS parallels (
		Id BIGSERIAL PRIMARY KEY,
		Name TEXT NOT NULL UNIQUE,
		MinGrade INTEGER NOT NULL,
		MaxGrade INTEGER NOT NULL,
		BestClassID INTEGER NOT NULL DEFAULT 0,
		ClassTotalRating INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}
	// ensure columns exist on older DBs (safe ALTER TABLE)
	if err := s.ensureColumn("parallels", "MinGrade", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("parallels", "MaxGrade", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("parallels", "BestClassID", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("parallels", "ClassTotalRating", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}

	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_parallels_min_max ON parallels(MinGrade, MaxGrade);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_parallels_name ON parallels(Name);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`
	CREATE TABLE IF NOT EXISTS parallel_classes (
		ParallelID INTEGER NOT NULL,
		ClassID INTEGER NOT NULL,
		PRIMARY KEY (ParallelID, ClassID),
		FOREIGN KEY (ParallelID) REFERENCES parallels(Id) ON DELETE CASCADE,
		FOREIGN KEY (ClassID) REFERENCES classes(Id) ON DELETE CASCADE
	);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_parallel_classes_parallel ON parallel_classes(ParallelID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_parallel_classes_class ON parallel_classes(ClassID);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initClassStorage() error {

	query := `
	CREATE TABLE IF NOT EXISTS classes (
		Id BIGSERIAL PRIMARY KEY,
		Name TEXT NOT NULL UNIQUE,
		Grade INTEGER NOT NULL DEFAULT 0,
		Letter TEXT NOT NULL DEFAULT '',
		TeacherLogin TEXT,
		Members JSONB NOT NULL DEFAULT '[]',
		FirstQuarterComplete INTEGER NOT NULL DEFAULT 0,
		SecondQuarterComplete INTEGER NOT NULL DEFAULT 0,
		ThirdQuarterComplete INTEGER NOT NULL DEFAULT 0,
		QuarterComplete INTEGER NOT NULL DEFAULT 0,
		UserTotalRating INTEGER NOT NULL DEFAULT 0,
		ClassTotalRating INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (TeacherLogin) REFERENCES users(Login) ON DELETE SET NULL
	);`

	if _, err := s.DB.Exec(query); err != nil {
		return err
	}

	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class ON users(Class);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(ClassID);`); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_classes_teacher_login ON classes(TeacherLogin);`); err != nil {
		return err
	}
	if err := s.ensureColumn("classes", "TotalRating", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}

	_, _ = s.DB.Exec(`ALTER TABLE classes ALTER COLUMN TotalRating SET DEFAULT 0;`)

	return syncAllClasses(s.DB)
}

func (s *Storage) ensureCurrentSchedulesSeeded() error {
	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM current_schedules`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err := s.DB.Exec(`
		INSERT INTO current_schedules
			(ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description)
		SELECT ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description
		FROM schedules
	`)
	return err
}

func (s *Storage) ensureColumn(table string, column string, definition string) error {
	var exists bool
	err := s.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = lower($1) AND column_name = lower($2)
		)
	`, table, column).Scan(&exists)
	
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return err
}

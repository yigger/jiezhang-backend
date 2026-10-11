-- Fresh development volumes only. The application never runs AutoMigrate.
SOURCE /docker-entrypoint-initdb.d/migrations/00000000_initial_schema.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261006_insights.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261007_project_details.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261007_fixed_cost_schedule.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261008_project_appearance.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261008_calendar_journal.sql;
SOURCE /docker-entrypoint-initdb.d/migrations/20261011_statement_import.sql;

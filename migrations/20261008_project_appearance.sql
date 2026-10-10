-- Project appearance is optional for compatibility with existing projects.
ALTER TABLE insight_projects
 ADD COLUMN icon VARCHAR(64) NULL,
 ADD COLUMN color VARCHAR(7) NULL;

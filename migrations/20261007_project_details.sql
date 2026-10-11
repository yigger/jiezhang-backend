-- Additive fields on analytics tables only; apply after 20261006_insights.sql.
ALTER TABLE insight_projects
 ADD COLUMN participant_ids JSON NULL,
 ADD COLUMN start_date DATE NULL,
 ADD COLUMN end_date DATE NULL;
ALTER TABLE insight_statement_annotations ADD COLUMN consumer_id BIGINT NULL;

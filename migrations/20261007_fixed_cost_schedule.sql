-- Apply after the analytics and project-details migrations.
-- Existing rules remain unscheduled until the user confirms their category and wallet.
ALTER TABLE insight_fixed_costs ADD COLUMN next_run_date DATE NULL;
CREATE TABLE insight_fixed_cost_runs (
 id BIGINT NOT NULL AUTO_INCREMENT,
 fixed_cost_id BIGINT NOT NULL,
 due_date DATE NOT NULL,
 statement_id BIGINT NOT NULL,
 created_at DATETIME NOT NULL,
 PRIMARY KEY (id),
 UNIQUE KEY idx_insight_fixed_cost_run (fixed_cost_id, due_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

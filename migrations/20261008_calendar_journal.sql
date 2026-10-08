-- Personal journals are isolated by user, account book and local calendar date.
CREATE TABLE calendar_journals (
 id BIGINT NOT NULL AUTO_INCREMENT,
 account_book_id BIGINT NOT NULL,
 user_id BIGINT NOT NULL,
 date VARCHAR(10) NOT NULL,
 mood VARCHAR(16) NOT NULL DEFAULT '',
 note VARCHAR(800) NOT NULL DEFAULT '',
 zero_expense TINYINT(1) NOT NULL DEFAULT 0,
 created_at DATETIME NOT NULL,
 updated_at DATETIME NOT NULL,
 PRIMARY KEY (id),
 UNIQUE KEY idx_calendar_journal_day (account_book_id, user_id, date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- NULL preserves the balance behavior of existing and ordinary statements.
-- FALSE identifies historical records, including after later edits.
-- Imported statements store SHA-256(source + ':' + original order ID) and a
-- separate fingerprint of the original bill contents. Ordinary rows keep NULL.
ALTER TABLE statements
 ADD COLUMN balance_applied TINYINT(1) NULL DEFAULT NULL,
 ADD COLUMN import_key CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL DEFAULT NULL,
 ADD COLUMN import_fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL DEFAULT NULL,
 ADD UNIQUE KEY idx_statements_import_key (account_book_id, user_id, import_key);

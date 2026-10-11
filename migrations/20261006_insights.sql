-- Additive analytics storage. Apply once before enabling the new write APIs.
-- Existing financial tables and values are not changed.

CREATE TABLE `insight_projects` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` bigint NOT NULL,
  `creator_id` bigint NOT NULL,
  `name` varchar(100) NOT NULL,
  `budget_cents` bigint NOT NULL,
  `archived` tinyint(1) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_insight_projects_book` (`account_book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `insight_fixed_costs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` bigint NOT NULL,
  `creator_id` bigint NOT NULL,
  `name` varchar(100) NOT NULL,
  `amount_cents` bigint NOT NULL,
  `category_id` bigint NOT NULL,
  `asset_id` bigint NOT NULL,
  `interval_months` int NOT NULL,
  `due_day` int NOT NULL,
  `candidate_key` varchar(255) NOT NULL,
  `active` tinyint(1) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_insight_fixed_costs_book` (`account_book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `insight_statement_annotations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` bigint NOT NULL,
  `creator_id` bigint NOT NULL,
  `statement_id` bigint NOT NULL,
  `project_id` bigint NULL,
  `payer_id` bigint NULL,
  `allocations` json NULL,
  `fixed_cost_id` bigint NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_insight_statement_annotations_book` (`account_book_id`),
  UNIQUE KEY `idx_insight_annotation_statement` (`statement_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `insight_portfolio_snapshots` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` bigint NOT NULL,
  `creator_id` bigint NOT NULL,
  `assets_cents` bigint NOT NULL,
  `liabilities_cents` bigint NOT NULL,
  `balances` json NOT NULL,
  `note` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_insight_portfolio_snapshots_book` (`account_book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `insight_merchant_aliases` (
 `id` bigint NOT NULL AUTO_INCREMENT,
 `account_book_id` bigint NOT NULL,
 `creator_id` bigint NOT NULL,
 `payee_id` bigint NOT NULL,
 `name` varchar(100) NOT NULL,
 `created_at` datetime NOT NULL,
 `updated_at` datetime NOT NULL,
 PRIMARY KEY (`id`),
 KEY `idx_insight_merchant_aliases_book` (`account_book_id`),
 UNIQUE KEY `idx_insight_merchant_aliases_payee` (`payee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

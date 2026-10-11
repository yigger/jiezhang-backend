-- Initial schema baseline for the 24 core tables, before the dated migrations.
-- Apply once to an empty database, then apply the remaining migrations in filename order.
-- Together the migration chain defines the complete current database structure.
-- Explicit DDL only: no credentials, user data, development seeds or AutoMigrate.
SET NAMES utf8mb4;

CREATE TABLE `account_books` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `account_type` int NULL DEFAULT 0,
  `name` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `description` varchar(255) NULL,
  `budget` decimal(12,2) NULL DEFAULT 0.00,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `account_book_collaborators` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` int NULL,
  `user_id` int NULL,
  `role` varchar(255) NULL DEFAULT 'member',
  `remark` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_account_book_collaborators_account_book_id_user_id` (`account_book_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `account_book_friend_refs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account_book_id` bigint NULL,
  `friend_id` bigint NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `index_account_book_friend_refs_on_account_book_id` (`account_book_id`),
  KEY `index_account_book_friend_refs_on_friend_id` (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `ar_internal_metadata` (
  `key` varchar(255) NOT NULL,
  `value` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `assets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NULL,
  `amount` decimal(12,2) NOT NULL DEFAULT 0.00,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `parent_id` int NULL DEFAULT 0,
  `type` varchar(255) NULL DEFAULT 'deposit',
  `lock` int NULL DEFAULT 0,
  `icon_path` varchar(255) NULL,
  `remark` text NULL,
  `creator_id` int NULL,
  `frequent` int NULL DEFAULT 0,
  `order` int NULL DEFAULT 0,
  `account_book_id` int NULL,
  PRIMARY KEY (`id`),
  KEY `index_assets_on_amount` (`amount`),
  KEY `index_assets_on_order` (`order`),
  KEY `index_assets_on_parent_id` (`parent_id`),
  KEY `index_assets_on_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `asset_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `type` int NULL,
  `from` int NULL,
  `to` int NULL,
  `amount` decimal(12,2) NULL,
  `residue` decimal(12,2) NULL,
  `description` text NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `account_book_id` int NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `asset_snapshots` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `statement_id` int NULL,
  `asset_id` int NULL,
  `year` int NULL,
  `month` int NULL,
  `day` int NULL,
  `date` date NULL,
  `before_value` decimal(12,2) NULL,
  `after_value` decimal(12,2) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `bonus_points_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `year` int NULL,
  `month` int NULL,
  `day` int NULL,
  `type` int NULL DEFAULT 0,
  `point` int NULL DEFAULT 0,
  `user_id` int NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `categories` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `name` varchar(255) NULL,
  `parent_id` int NOT NULL DEFAULT 0,
  `order` int NOT NULL DEFAULT 0,
  `icon_path` varchar(255) NULL,
  `color` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `type` varchar(255) NULL,
  `lock` int NULL DEFAULT 0,
  `budget` decimal(12,2) NULL DEFAULT 0.00,
  `frequent` int NULL DEFAULT 0,
  `is_mess` int NULL DEFAULT 0,
  `account_book_id` int NULL,
  `is_system` tinyint(1) NULL DEFAULT 0,
  `special_type` varchar(255) NULL,
  PRIMARY KEY (`id`),
  KEY `index_categories_on_order` (`order`),
  KEY `index_categories_on_parent_id` (`parent_id`),
  KEY `index_categories_on_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `error_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `message` json NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `feedbacks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `content` text NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `type` int NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `friends` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `name` varchar(255) NULL,
  `target_id` int NULL,
  `remark` varchar(255) NULL,
  `role_name` varchar(255) NULL,
  `active` int NULL DEFAULT 0,
  `already_bind` tinyint(1) NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `friend_applies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `source_id` int NULL,
  `target_id` int NULL,
  `remark` varchar(255) NULL,
  `status` int NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NULL,
  `from_user_id` int NULL,
  `target_id` int NULL,
  `target_type` int NULL,
  `content` text NULL,
  `content_type` varchar(255) NULL DEFAULT 'md',
  `avatar_url` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `already_read` int NULL DEFAULT 0,
  `page_url` varchar(255) NULL,
  `sub_title` text NULL,
  `date` datetime NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `month_charts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `year` int NULL,
  `month` int NULL,
  `dashboard` json NULL,
  `expend_compare` json NULL,
  `day_avg` json NULL,
  `week_avg` json NULL,
  `month_surplus` json NULL,
  `budget_used` json NULL,
  `asset_total` json NULL,
  `month_last_10` json NULL,
  `begin_text` text NULL,
  `end_text` text NULL,
  `cover` text NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `operate_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `year` int NULL,
  `month` int NULL,
  `day` int NULL,
  `page_route` varchar(512) NULL,
  `session_key` varchar(255) NULL,
  `ip` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `payees` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NULL,
  `account_book_id` bigint NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `user_id` int NULL,
  PRIMARY KEY (`id`),
  KEY `index_payees_on_account_book_id` (`account_book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `pre_orders` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` int NULL,
  `creator_id` int NULL,
  `name` varchar(255) NULL,
  `amount` decimal(12,2) NULL,
  `state` varchar(255) NULL DEFAULT 'pending',
  `remark` text NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `address` varchar(255) NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `recommends` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `share_ticket` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `schema_migrations` (
  `version` varchar(255) NOT NULL,
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `statements` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `category_id` int NULL,
  `asset_id` int NOT NULL,
  `amount` decimal(12,2) NULL,
  `refund` decimal(12,2) NULL DEFAULT 0.00,
  `type` varchar(255) NOT NULL,
  `description` text NULL,
  `year` int NULL,
  `month` int NULL,
  `day` int NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `time` time NULL,
  `residue` decimal(12,2) NULL DEFAULT 0.00,
  `location` text NULL,
  `nation` varchar(255) NULL,
  `province` varchar(255) NULL,
  `city` varchar(255) NULL,
  `district` varchar(255) NULL,
  `street` varchar(255) NULL,
  `target_asset_id` int NULL,
  `account_book_id` int NULL,
  `ref` varchar(255) NULL,
  `ref_at` datetime NULL,
  `mood` varchar(255) NULL,
  `payee_id` int NULL,
  `target_object` varchar(255) NULL,
  PRIMARY KEY (`id`),
  KEY `index_statements_on_type` (`type`),
  KEY `index_statements_on_user_id_and_asset_id` (`user_id`, `asset_id`),
  KEY `index_statements_on_user_id_and_category_id` (`user_id`, `category_id`),
  KEY `index_statements_on_user_id_and_type` (`user_id`, `type`),
  KEY `index_statements_on_year_and_month_and_day_and_time` (`year`, `month`, `day`, `time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NULL DEFAULT '',
  `theme_id` int NULL DEFAULT 0,
  `openid` varchar(255) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `nickname` varchar(255) NULL,
  `language` varchar(255) NULL,
  `city` varchar(255) NULL,
  `province` varchar(255) NULL,
  `avatar_url` varchar(512) NULL,
  `country` varchar(255) NULL,
  `session_key` varchar(255) NULL,
  `gender` int NULL,
  `uid` int NOT NULL DEFAULT 0,
  `third_session` varchar(255) NULL,
  `phone` varchar(255) NULL,
  `budget` decimal(12,2) NULL DEFAULT 0.00,
  `bg_avatar_url` varchar(255) NULL,
  `bonus_points` int NULL DEFAULT 0,
  `header_position_1` varchar(255) NULL,
  `header_position_2` varchar(255) NULL,
  `header_position_3` varchar(255) NULL,
  `bg_avatar_id` int NULL,
  `remind` int NULL DEFAULT 0,
  `hidden_asset_money` tinyint(1) NULL DEFAULT 0,
  `already_login` tinyint(1) NULL DEFAULT 0,
  `admin` tinyint(1) NULL DEFAULT 0,
  `password` varchar(255) NULL,
  `salt` varchar(255) NULL,
  `account_book_id` int NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_open_id` (`openid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `user_assets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NULL,
  `path` varchar(255) NULL,
  `type` varchar(255) NULL,
  `imageable_type` varchar(255) NULL,
  `imageable_id` bigint NULL,
  `score` int NULL DEFAULT 0,
  `locked` int NULL DEFAULT 0,
  `system` int NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `index_user_assets_on_imageable_type_and_imageable_id` (`imageable_type`(191), `imageable_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `users_assets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NULL,
  `asset_id` varchar(255) NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `account_book_id` int NULL,
  `assign_type` int NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


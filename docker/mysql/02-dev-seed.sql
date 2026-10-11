-- Synthetic local development identities and empty wallets. No real user data.
SET NAMES utf8mb4;
START TRANSACTION;

INSERT INTO users (id, openid, nickname, uid, account_book_id, created_at, updated_at)
VALUES (1, 'docker-dev-user', '开发用户', 1, 1, NOW(3), NOW(3));

INSERT INTO account_books (id, user_id, account_type, name, description, created_at, updated_at)
VALUES (1, 1, 0, '开发账本', 'Docker 本地开发环境', NOW(), NOW());

INSERT INTO account_book_collaborators (account_book_id, user_id, role, created_at, updated_at)
VALUES (1, 1, 'admin', NOW(), NOW());

INSERT INTO categories (id, user_id, name, parent_id, type, account_book_id, is_system, created_at, updated_at)
VALUES
 (1, 1, '伙食餐饮', 0, 'expend', 1, 0, NOW(), NOW()),
 (2, 1, '早餐', 1, 'expend', 1, 0, NOW(), NOW()),
 (3, 1, '工作收入', 0, 'income', 1, 0, NOW(), NOW()),
 (4, 1, '工资', 3, 'income', 1, 0, NOW(), NOW()),
 (5, 1, '其他杂项', 0, 'expend', 1, 0, NOW(), NOW()),
 (6, 1, '其他支出', 5, 'expend', 1, 0, NOW(), NOW());

INSERT INTO categories (id, user_id, name, parent_id, type, special_type, account_book_id, is_system, created_at, updated_at)
VALUES
 (100, 1, '转账', 0, 'transfer', 'transfer', 1, 1, NOW(), NOW()),
 (101, 1, '还款', 0, 'repayment', 'repayment', 1, 1, NOW(), NOW()),
 (102, 1, '借入', 0, 'loan_in', 'loan_in', 1, 1, NOW(), NOW()),
 (103, 1, '借出', 0, 'loan_out', 'loan_out', 1, 1, NOW(), NOW()),
 (104, 1, '报销', 0, 'reimburse', 'reimburse', 1, 1, NOW(), NOW()),
 (105, 1, '代付', 0, 'payment_proxy', 'payment_proxy', 1, 1, NOW(), NOW());

INSERT INTO assets (id, name, amount, parent_id, type, creator_id, account_book_id, created_at, updated_at)
VALUES
 (1, '储蓄账户', 0, 0, 'deposit', 1, 1, NOW(), NOW()),
 (2, '现金', 0, 1, 'deposit', 1, 1, NOW(), NOW()),
 (3, '虚拟账户', 0, 0, 'deposit', 1, 1, NOW(), NOW()),
 (4, '支付宝', 0, 3, 'deposit', 1, 1, NOW(), NOW()),
 (5, '微信', 0, 3, 'deposit', 1, 1, NOW(), NOW()),
 (6, '负债账户', 0, 0, 'debt', 1, 1, NOW(), NOW()),
 (7, '信用卡', 0, 6, 'debt', 1, 1, NOW(), NOW());

COMMIT;

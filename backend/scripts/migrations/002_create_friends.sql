CREATE TABLE IF NOT EXISTS `friend_requests` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT NOT NULL,
    `friend_id` BIGINT NOT NULL,
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/accepted/blocked',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX `idx_user_friend` (`user_id`, `friend_id`),
    INDEX `idx_friend_id` (`friend_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

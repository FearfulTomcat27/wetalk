CREATE TABLE IF NOT EXISTS `friendships` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user1_id` BIGINT NOT NULL COMMENT '较小 ID',
    `user2_id` BIGINT NOT NULL COMMENT '较大 ID (user1_id < user2_id)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user1_user2` (`user1_id`, `user2_id`),
    INDEX `idx_user1` (`user1_id`),
    INDEX `idx_user2` (`user2_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

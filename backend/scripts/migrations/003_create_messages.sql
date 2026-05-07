CREATE TABLE IF NOT EXISTS `messages` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `sender_id` BIGINT NOT NULL,
    `receiver_id` BIGINT NOT NULL,
    `content` TEXT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_sender_receiver` (`sender_id`, `receiver_id`),
    INDEX `idx_receiver_sender` (`receiver_id`, `sender_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

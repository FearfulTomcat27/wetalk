CREATE TABLE IF NOT EXISTS `messages` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `sender_id` BIGINT NOT NULL,
    `receiver_id` BIGINT NOT NULL,
    `content` TEXT NOT NULL,
    `content_type` VARCHAR(16) NOT NULL DEFAULT 'text' COMMENT 'text/image/file',
    `status` VARCHAR(16) NOT NULL DEFAULT 'sent' COMMENT 'sent/delivered/read',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_sender_receiver` (`sender_id`, `receiver_id`),
    INDEX `idx_receiver_sender` (`receiver_id`, `sender_id`),
    INDEX `idx_created_at` (`created_at`),
    CONSTRAINT `fk_messages_sender` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_messages_receiver` FOREIGN KEY (`receiver_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

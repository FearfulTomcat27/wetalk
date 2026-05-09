CREATE TABLE `chats` (
                         `id` bigint NOT NULL AUTO_INCREMENT,
                         `chat_type` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'private' COMMENT 'private / group',
                         `chat_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
                         `last_message_id` bigint DEFAULT NULL,
                         `last_message_text` text COLLATE utf8mb4_unicode_ci,
                         `last_message_time` datetime DEFAULT NULL,
                         `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                         PRIMARY KEY (`id`),
                         KEY `fk_chats_last_message` (`last_message_id`),
                         CONSTRAINT `fk_chats_last_message` FOREIGN KEY (`last_message_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
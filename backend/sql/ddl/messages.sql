CREATE TABLE `messages` (
                            `id` bigint NOT NULL AUTO_INCREMENT,
                            `chat_id` bigint NOT NULL,
                            `sender_id` bigint NOT NULL,
                            `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
                            `content_type` enum('text','image','file') COLLATE utf8mb4_unicode_ci DEFAULT 'text',
                            `status` enum('sent','delivered','read','revoked') COLLATE utf8mb4_unicode_ci DEFAULT 'sent',
                            `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
                            `quote_id` bigint DEFAULT NULL,
                            PRIMARY KEY (`id`),
                            KEY `idx_chat_id` (`chat_id`),
                            KEY `idx_chat_created` (`chat_id`,`created_at`),
                            KEY `idx_quote_id` (`quote_id`),
                            CONSTRAINT `fk_messages_chat` FOREIGN KEY (`chat_id`) REFERENCES `chats` (`id`) ON DELETE CASCADE,
                            CONSTRAINT `fk_messages_quote` FOREIGN KEY (`quote_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
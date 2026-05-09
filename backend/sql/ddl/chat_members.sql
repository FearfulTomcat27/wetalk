CREATE TABLE `chat_members` (
                                `id` bigint NOT NULL AUTO_INCREMENT,
                                `chat_id` bigint NOT NULL,
                                `user_id` bigint NOT NULL,
                                `joined_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                PRIMARY KEY (`id`),
                                UNIQUE KEY `uk_chat_user` (`chat_id`,`user_id`),
                                KEY `idx_user_id` (`user_id`),
                                CONSTRAINT `fk_chat_members_chat` FOREIGN KEY (`chat_id`) REFERENCES `chats` (`id`) ON DELETE CASCADE,
                                CONSTRAINT `fk_chat_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
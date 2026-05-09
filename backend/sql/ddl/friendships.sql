CREATE TABLE `friendships` (
                               `id` bigint NOT NULL AUTO_INCREMENT,
                               `chat_id` bigint NOT NULL,
                               `user1_id` bigint NOT NULL COMMENT '较小 ID',
                               `user2_id` bigint NOT NULL COMMENT '较大 ID (user1_id < user2_id)',
                               `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `uk_user1_user2` (`user1_id`,`user2_id`),
                               KEY `idx_user1` (`user1_id`),
                               KEY `idx_user2` (`user2_id`),
                               KEY `idx_chat_id` (`chat_id`),
                               CONSTRAINT `fk_friendships_chat` FOREIGN KEY (`chat_id`) REFERENCES `chats` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
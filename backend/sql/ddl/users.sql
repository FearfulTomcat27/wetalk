CREATE TABLE `users` (
                         `id` bigint NOT NULL AUTO_INCREMENT,
                         `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
                         `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
                         `nickname` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
                         `avatar` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
                         `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
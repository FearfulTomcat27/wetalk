CREATE TABLE `file_metadata` (
                                 `id` bigint NOT NULL AUTO_INCREMENT,
                                 `message_id` bigint NOT NULL COMMENT '关联消息ID',
                                 `url` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件OSS存储URL',
                                 `original_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件原始名称',
                                 `file_size` bigint DEFAULT NULL COMMENT '文件大小(字节)',
                                 `mime_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件MIME类型',
                                 `width` int DEFAULT NULL COMMENT '图片宽度(像素)',
                                 `height` int DEFAULT NULL COMMENT '图片高度(像素)',
                                 `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_message_id` (`message_id`),
                                 CONSTRAINT `fk_file_metadata_message` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
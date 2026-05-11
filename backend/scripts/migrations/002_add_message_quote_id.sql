ALTER TABLE `messages` ADD COLUMN `quote_id` bigint DEFAULT NULL AFTER `created_at`;
ALTER TABLE `messages` ADD INDEX `idx_quote_id` (`quote_id`);
ALTER TABLE `messages` ADD CONSTRAINT `fk_messages_quote` FOREIGN KEY (`quote_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL;

-- 002_index_optimizations.sql
-- 索引优化：修复慢查询、优化未读统计和待处理好友请求查询
--
-- P0: messages   — 新增 (chat_id, status) 复合索引 + 删除冗余 idx_chat_id
-- P1: friend_requests — 用 (friend_id, status, created_at) 覆盖索引替换两个单列索引
-- P2: friend_requests — 复合索引已覆盖，删除冗余单列索引

-- ========== P0: messages 表 ==========

-- 复合索引 (chat_id, status) 让 MarkAsRead 的 UPDATE 和未读计数查询
-- 在索引层直接跳过已读行，避免逐行回表扫描 chat 内所有消息
-- 覆盖查询: WHERE chat_id = ? AND status != 'read'
-- 覆盖查询: WHERE chat_id = ? AND sender_id != ? AND status != 'read' (回表后 filter sender_id)
CREATE INDEX `idx_messages_chat_status` ON `messages` (`chat_id`, `status`);

-- idx_chat_id 是 idx_chat_created(chat_id, created_at) 的左前缀，功能完全被覆盖
-- 删除冗余索引可减少 B 树体积，降低索引维护开销
DROP INDEX `idx_chat_id` ON `messages`;

-- ========== P1 + P2: friend_requests 表 ==========

-- 复合索引 (friend_id, status, created_at) 覆盖 WHERE + ORDER BY
-- MySQL 在单查询中只能为每张表使用一个索引，两个单列索引无法同时生效
-- 覆盖查询: WHERE friend_id = ? AND status = ? ORDER BY created_at DESC (FindPendingByUserID)
CREATE INDEX `idx_friend_id_status_created` ON `friend_requests` (`friend_id`, `status`, `created_at`);

-- 复合索引已覆盖 friend_id 和 status 的查询场景，删除冗余单列索引
-- idx_friend_id 和 idx_status 不再有独立存在的价值
DROP INDEX `idx_friend_id` ON `friend_requests`;
DROP INDEX `idx_status` ON `friend_requests`;

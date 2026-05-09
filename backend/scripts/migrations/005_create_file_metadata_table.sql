-- 006: 创建 file_metadata 表，删除 messages 表的文件元数据字段

CREATE TABLE IF NOT EXISTS file_metadata (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id BIGINT NOT NULL COMMENT '关联消息ID',
    url TEXT NOT NULL COMMENT '文件OSS存储URL',
    original_name VARCHAR(255) DEFAULT NULL COMMENT '文件原始名称',
    file_size BIGINT DEFAULT NULL COMMENT '文件大小(字节)',
    mime_type VARCHAR(100) DEFAULT NULL COMMENT '文件MIME类型',
    width INT DEFAULT NULL COMMENT '图片宽度(像素)',
    height INT DEFAULT NULL COMMENT '图片高度(像素)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_message_id (message_id),
    CONSTRAINT fk_file_metadata_message FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

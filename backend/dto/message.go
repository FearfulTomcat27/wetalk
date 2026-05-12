package dto

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"wetalk/db"
	"wetalk/model"
)

// messageRepo 消息数据仓库（MongoDB 实现）
type messageRepo struct{}

// Message 消息仓库实例
var Message = &messageRepo{}

// UnreadCount 单个聊天的未读数
type UnreadCount struct {
	ChatID int64 `json:"chat_id"`
	Count  int64 `json:"count"`
}

// nextMsgID 从 MongoDB counters 集合获取自增 msg_id
func nextMsgID(ctx context.Context) (int64, error) {
	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := db.CountersCollection().FindOneAndUpdate(
		ctx,
		bson.M{"_id": "message_id"},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

// Create 创建消息（含嵌入式 file_metadata 和引用消息内容预填充）
func (r *messageRepo) Create(senderID, chatID int64, content string, contentType string, quoteID *int64, fileMeta *model.FileMetadataPayload) (*model.MessageResponse, error) {
	ctx := context.Background()

	msgID, err := nextMsgID(ctx)
	if err != nil {
		return nil, err
	}

	// 如果引用消息，预填充被引用内容
	var quotedContent *string
	var quotedSenderID *int64
	var quotedSenderName *string
	var quotedContentType *string
	var quotedFileMeta *model.FileMetadataEmbed
	if quoteID != nil {
		quoted, err := r.GetByID(*quoteID)
		if err != nil {
			return nil, err
		}
		if quoted != nil {
			quotedContent = &quoted.Content
			quotedSenderID = &quoted.SenderID
			quotedContentType = &quoted.ContentType
			if quoted.FileMetadata != nil {
				quotedFileMeta = &model.FileMetadataEmbed{
					URL:          quoted.FileMetadata.URL,
					OriginalName: quoted.FileMetadata.OriginalName,
					FileSize:     quoted.FileMetadata.FileSize,
					MimeType:     quoted.FileMetadata.MimeType,
					Width:        quoted.FileMetadata.Width,
					Height:       quoted.FileMetadata.Height,
				}
			}
			// 查询发送者用户名
			var username string
			if err := db.DB.Table("users").Select("username").Where("id = ?", quoted.SenderID).Scan(&username).Error; err == nil && username != "" {
				quotedSenderName = &username
			}
		}
	}

	// 构建嵌入式文件元数据
	var embedMeta *model.FileMetadataEmbed
	if fileMeta != nil {
		embedMeta = &model.FileMetadataEmbed{
			URL:          fileMeta.URL,
			OriginalName: fileMeta.OriginalName,
			FileSize:     fileMeta.FileSize,
			MimeType:     fileMeta.MimeType,
			Width:        fileMeta.Width,
			Height:       fileMeta.Height,
		}
	}

	now := time.Now()
	doc := model.MessageDoc{
		MsgID:             msgID,
		ChatID:            chatID,
		SenderID:          senderID,
		Content:           content,
		ContentType:       contentType,
		Status:            model.MessageStatusSent,
		QuoteMessageID:    quoteID,
		QuotedContent:     quotedContent,
		QuotedSenderID:    quotedSenderID,
		QuotedSenderName:  quotedSenderName,
		QuotedContentType: quotedContentType,
		QuotedFileMeta:    quotedFileMeta,
		FileMetadata:      embedMeta,
		CreatedAt:         now,
	}

	if _, err := db.MsgCollection().InsertOne(ctx, &doc); err != nil {
		return nil, err
	}

	resp := doc.ToResponse()
	return &resp, nil
}

// docToMessage 将 MessageDoc 转换为 Message（兼容旧接口）
func docToMessage(doc *model.MessageDoc) *model.Message {
	if doc == nil {
		return nil
	}
	m := &model.Message{
		ID:             doc.MsgID,
		ChatID:         doc.ChatID,
		SenderID:       doc.SenderID,
		Content:        doc.Content,
		ContentType:    doc.ContentType,
		Status:         doc.Status,
		QuoteMessageID: doc.QuoteMessageID,
		CreatedAt:      doc.CreatedAt,
	}
	if doc.FileMetadata != nil {
		m.FileMetadata = &model.FileMetadata{
			URL:          doc.FileMetadata.URL,
			OriginalName: doc.FileMetadata.OriginalName,
			FileSize:     doc.FileMetadata.FileSize,
			MimeType:     doc.FileMetadata.MimeType,
			Width:        doc.FileMetadata.Width,
			Height:       doc.FileMetadata.Height,
		}
	}
	return m
}

// docToMessageResponse 将 MessageDoc 转换为 MessageResponse
func docToMessageResponse(doc model.MessageDoc) model.MessageResponse {
	return doc.ToResponse()
}

// GetByID 根据 msg_id 查询单条消息
func (r *messageRepo) GetByID(msgID int64) (*model.Message, error) {
	ctx := context.Background()
	var doc model.MessageDoc
	err := db.MsgCollection().FindOne(ctx, bson.M{"msg_id": msgID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return docToMessage(&doc), nil
}

// ListByIDs 批量查询消息（用于引用消息查找）
func (r *messageRepo) ListByIDs(msgIDs []int64) ([]model.Message, error) {
	if len(msgIDs) == 0 {
		return nil, nil
	}
	ctx := context.Background()
	cursor, err := db.MsgCollection().Find(ctx, bson.M{"msg_id": bson.M{"$in": msgIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []model.MessageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	messages := make([]model.Message, len(docs))
	for i, doc := range docs {
		m := docToMessage(&doc)
		if m != nil {
			messages[i] = *m
		}
	}
	return messages, nil
}

// MarkAsRead 将聊天中对方发来的未读消息标记为已读
func (r *messageRepo) MarkAsRead(chatID, currentUserID int64) error {
	ctx := context.Background()
	_, err := db.MsgCollection().UpdateMany(
		ctx,
		bson.M{
			"chat_id":   chatID,
			"sender_id": bson.M{"$ne": currentUserID},
			"status":    bson.M{"$ne": model.MessageStatusRead},
		},
		bson.M{"$set": bson.M{"status": model.MessageStatusRead}},
	)
	return err
}

// ListByChat 获取聊天消息记录（分页，按时间倒序）
func (r *messageRepo) ListByChat(chatID int64, offset, limit int) ([]model.MessageResponse, error) {
	ctx := context.Background()

	findOpts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.MsgCollection().Find(ctx, bson.M{"chat_id": chatID}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []model.MessageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	responses := make([]model.MessageResponse, len(docs))
	for i, doc := range docs {
		responses[i] = docToMessageResponse(doc)
	}
	return responses, nil
}

// CountUnread 统计聊天中当前用户未读的消息数
func (r *messageRepo) CountUnread(chatID, currentUserID int64) (int64, error) {
	ctx := context.Background()
	return db.MsgCollection().CountDocuments(ctx, bson.M{
		"chat_id":   chatID,
		"sender_id": bson.M{"$ne": currentUserID},
		"status":    bson.M{"$ne": model.MessageStatusRead},
	})
}

// GetUnreadCounts 查询当前用户在所有聊天中的未读消息数（供侧边栏角标使用）
func (r *messageRepo) GetUnreadCounts(userID int64) ([]UnreadCount, error) {
	ctx := context.Background()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{
			"sender_id": bson.M{"$ne": userID},
			"status":    bson.M{"$ne": model.MessageStatusRead},
		}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$chat_id",
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$project", Value: bson.M{
			"_id":     0,
			"chat_id": "$_id",
			"count":   1,
		}}},
	}

	cursor, err := db.MsgCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ChatID int64 `bson:"chat_id"`
		Count  int64 `bson:"count"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	counts := make([]UnreadCount, len(results))
	for i, r := range results {
		counts[i] = UnreadCount{
			ChatID: r.ChatID,
			Count:  r.Count,
		}
	}
	return counts, nil
}

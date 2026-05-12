package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
	"wetalk/types"
)

func newTestMessageService(msgRepo *MockMessageRepo, chatSvc *ChatService, cache *MockCache) *MessageService {
	return NewMessageService(msgRepo, chatSvc, cache)
}

func TestMessageService_SendMessage_EmptyContent(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatSvc := newTestChatService(new(MockChatRepo), cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	_, err := svc.SendMessage(1, model.SendMessageRequest{
		ChatID:  1,
		Content: "",
	})
	assert.ErrorIs(t, err, types.ErrInvalidParam)
}

func TestMessageService_SendMessage_NotMember(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	chatRepo.On("GetMemberIDs", int64(1)).Return([]int64{2, 3}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{2, 3}, mock.AnythingOfType("time.Duration")).Return(nil)

	_, err := svc.SendMessage(1, model.SendMessageRequest{
		ChatID:  1,
		Content: "hello",
	})
	assert.ErrorIs(t, err, types.ErrNotChatMember)
}

func TestMessageService_SendMessage_QuoteNotFound(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	quoteID := int64(999)

	chatRepo.On("GetMemberIDs", int64(1)).Return([]int64{1, 2}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{1, 2}, mock.AnythingOfType("time.Duration")).Return(nil)
	msgRepo.On("GetByID", int64(999)).Return(nil, nil)

	_, err := svc.SendMessage(1, model.SendMessageRequest{
		ChatID:         1,
		Content:        "reply",
		QuoteMessageID: &quoteID,
	})
	assert.ErrorIs(t, err, types.ErrQuotedMessageNotFound)
}

func TestMessageService_SendMessage_Success(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	now := time.Now()

	chatRepo.On("GetMemberIDs", int64(1)).Return([]int64{1, 2}, nil)
	cache.On("Get", "chat_members:1", mock.Anything).Return(false, nil)
	cache.On("Set", "chat_members:1", []int64{1, 2}, mock.AnythingOfType("time.Duration")).Return(nil)
	msgRepo.On("Create", int64(1), int64(1), "hello", model.ContentTypeText, (*int64)(nil), (*model.FileMetadataPayload)(nil)).
		Return(&model.MessageResponse{
			ID: 100, ChatID: 1, SenderID: 1, Content: "hello",
			ContentType: model.ContentTypeText, Status: model.MessageStatusSent,
			CreatedAt: now,
		}, nil)
	chatRepo.On("UpdateLastMessage", int64(1), int64(100), "hello", model.ContentTypeText, now).Return(nil)

	resp, err := svc.SendMessage(1, model.SendMessageRequest{
		ChatID:  1,
		Content: "hello",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(100), resp.ID)
	assert.Equal(t, "hello", resp.Content)
}

func TestMessageService_MarkAsRead_Success(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	msgRepo.On("MarkAsRead", int64(1), int64(1)).Return(nil)
	cache.On("Del", []string{"unread:1"}).Return(nil)

	err := svc.MarkAsRead(1, 1)
	require.NoError(t, err)

	msgRepo.AssertExpectations(t)
}

func TestMessageService_GetUnreadCounts_CacheHit(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	cached := []dto.UnreadCount{{ChatID: 1, Count: 5}}
	cache.On("Get", "unread:1", mock.Anything).Return(true, nil, cached)

	counts, err := svc.GetUnreadCounts(1)
	require.NoError(t, err)
	assert.Len(t, counts, 1)

	msgRepo.AssertNotCalled(t, "GetUnreadCounts")
	cache.AssertExpectations(t)
}

func TestMessageService_GetConversation_DefaultsLimit(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	cache := new(MockCache)
	chatRepo := new(MockChatRepo)
	chatSvc := newTestChatService(chatRepo, cache)
	svc := newTestMessageService(msgRepo, chatSvc, cache)

	msgRepo.On("ListByChat", int64(1), 0, 50).Return([]model.MessageResponse{}, nil)

	resp, err := svc.GetConversation(1, 0, 0)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	msgRepo.AssertExpectations(t)
}

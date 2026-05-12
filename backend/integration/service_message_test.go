//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
	"wetalk/service"
)

func TestServiceMessage_SendAndGetConversation(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	// Send message
	resp, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
		ChatID:  chat.ID,
		Content: "Hello Bob!",
	})
	require.NoError(t, err)
	assert.NotZero(t, resp.ID)
	assert.Equal(t, "Hello Bob!", resp.Content)
	assert.Equal(t, model.ContentTypeText, resp.ContentType)

	// Get conversation
	msgs, err := msgSvc.GetConversation(chat.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
}

func TestServiceMessage_Send_NotMember(t *testing.T) {
	resetDB(t)

	_, _, chat := CreateFriendPair(t)
	charles := CreateUser(t, "charles_"+t.Name())

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	_, err := msgSvc.SendMessage(charles.ID, model.SendMessageRequest{
		ChatID:  chat.ID,
		Content: "Hello",
	})
	require.Error(t, err)
	assert.Equal(t, "[not_chat_member] 不是聊天成员", err.Error())
}

func TestServiceMessage_Send_EmptyContent(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	_, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
		ChatID:  chat.ID,
		Content: "",
	})
	require.Error(t, err)
	assert.Equal(t, "[invalid_param] 请求参数错误", err.Error())
}

func TestServiceMessage_SendWithQuote(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	// Bob sends original
	original, err := msgSvc.SendMessage(bob.ID, model.SendMessageRequest{
		ChatID:  chat.ID,
		Content: "Original message",
	})
	require.NoError(t, err)

	// Alice quotes it
	quoted, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
		ChatID:         chat.ID,
		Content:        "Replying",
		QuoteMessageID: &original.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, quoted.QuotedContent)
	assert.Equal(t, "Original message", *quoted.QuotedContent)
}

func TestServiceMessage_SendWithQuote_WrongChat(t *testing.T) {
	resetDB(t)

	alice, bob, chat1 := CreateFriendPair(t)
	charles := CreateUser(t, "charles_"+t.Name())
	chat2, _ := CreateFriendship(t, bob, charles)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	// Create message in chat1
	original, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
		ChatID:  chat1.ID,
		Content: "Secret message",
	})
	require.NoError(t, err)

	// Try to quote from chat1 in chat2
	_, err = msgSvc.SendMessage(bob.ID, model.SendMessageRequest{
		ChatID:         chat2.ID,
		Content:        "Reply with wrong quote",
		QuoteMessageID: &original.ID,
	})
	require.Error(t, err)
	assert.Equal(t, "[quoted_message_not_found] 引用的消息不存在", err.Error())
}

func TestServiceMessage_MarkAsRead(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	// Alice sends 3 messages
	for i := 0; i < 3; i++ {
		_, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
			ChatID:  chat.ID,
			Content: "Msg",
		})
		require.NoError(t, err)
	}

	// Get unread counts (Bob should have 3 unread)
	counts, err := msgSvc.GetUnreadCounts(bob.ID)
	require.NoError(t, err)
	require.Len(t, counts, 1)
	assert.Equal(t, int64(3), counts[0].Count)

	// Bob marks as read
	err = msgSvc.MarkAsRead(bob.ID, chat.ID)
	require.NoError(t, err)

	// Unread should be 0 now
	counts, err = msgSvc.GetUnreadCounts(bob.ID)
	require.NoError(t, err)
	assert.Len(t, counts, 0)
}

func TestServiceMessage_GetConversation_Pagination(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	chatSvc := service.NewChatService(dto.Chat, cache)
	msgSvc := service.NewMessageService(dto.Message, chatSvc, cache)

	// Send 10 messages
	for i := 0; i < 10; i++ {
		_, err := msgSvc.SendMessage(alice.ID, model.SendMessageRequest{
			ChatID:  chat.ID,
			Content: string(rune('A' + i)),
		})
		require.NoError(t, err)
	}

	// Get page 1 (limit 3) — newest first
	msgs, err := msgSvc.GetConversation(chat.ID, 0, 3)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)

	// Get page 2
	msgs2, err := msgSvc.GetConversation(chat.ID, 3, 3)
	require.NoError(t, err)
	assert.Len(t, msgs2, 3)

	// Ensure no overlap (newest first)
	assert.NotEqual(t, msgs[0].ID, msgs2[0].ID)
}

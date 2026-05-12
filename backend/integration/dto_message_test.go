//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/dto"
	"wetalk/model"
)

func TestDTOMessage_CreateAndGetByID(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	// Create a text message
	msg, err := dto.Message.Create(alice.ID, chat.ID, "Hello Bob!", model.ContentTypeText, nil, nil)
	require.NoError(t, err)
	require.NotZero(t, msg.ID)
	assert.Equal(t, "Hello Bob!", msg.Content)
	assert.Equal(t, model.ContentTypeText, msg.ContentType)
	assert.Equal(t, model.MessageStatusSent, msg.Status)
	assert.Equal(t, alice.ID, msg.SenderID)

	// GetByID
	found, err := dto.Message.GetByID(msg.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, msg.Content, found.Content)
}

func TestDTOMessage_CreateWithFileMetadata(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	fileMeta := &model.FileMetadataPayload{
		URL:          "https://oss.example.com/test.pdf",
		OriginalName: "test.pdf",
		FileSize:     1024,
		MimeType:     "application/pdf",
	}

	msg, err := dto.Message.Create(alice.ID, chat.ID, "Check this file", model.ContentTypeFile, nil, fileMeta)
	require.NoError(t, err)
	require.NotNil(t, msg.FileMetadata)
	assert.Equal(t, "test.pdf", msg.FileMetadata.OriginalName)
	assert.Equal(t, int64(1024), msg.FileMetadata.FileSize)
	assert.Equal(t, "application/pdf", msg.FileMetadata.MimeType)
}

func TestDTOMessage_CreateWithQuote(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	// Create original message
	original, err := dto.Message.Create(bob.ID, chat.ID, "Original message", model.ContentTypeText, nil, nil)
	require.NoError(t, err)

	// Create quoted message
	quoted, err := dto.Message.Create(alice.ID, chat.ID, "Replying to you", model.ContentTypeText, &original.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, quoted.QuotedContent)
	assert.Equal(t, "Original message", *quoted.QuotedContent)
}

func TestDTOMessage_ListByChat(t *testing.T) {
	resetDB(t)

	alice, _, chat := CreateFriendPair(t)

	// Create multiple messages
	for i := 0; i < 5; i++ {
		_, err := dto.Message.Create(alice.ID, chat.ID, "Msg "+string(rune('0'+i)), model.ContentTypeText, nil, nil)
		require.NoError(t, err)
	}

	// List in reverse chronological order
	msgs, err := dto.Message.ListByChat(chat.ID, 1, 0, 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 5)

	// Verify ordering (newest first)
	for i := 1; i < len(msgs); i++ {
		assert.True(t, msgs[i-1].CreatedAt.Unix() >= msgs[i].CreatedAt.Unix(),
			"messages should be in descending order by time")
	}

	// Test pagination
	msgs2, err := dto.Message.ListByChat(chat.ID, 1, 0, 2)
	require.NoError(t, err)
	assert.Len(t, msgs2, 2)
}

func TestDTOMessage_MarkAsReadAndCountUnread(t *testing.T) {
	resetDB(t)

	alice, bob, chat := CreateFriendPair(t)

	// Alice sends 3 messages to Bob
	for i := 0; i < 3; i++ {
		_, err := dto.Message.Create(alice.ID, chat.ID, "Hello", model.ContentTypeText, nil, nil)
		require.NoError(t, err)
	}

	// Bob's unread count should be 3
	count, err := dto.Message.CountUnread(chat.ID, bob.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Alice's unread count should be 0 (she sent them)
	countAlice, err := dto.Message.CountUnread(chat.ID, alice.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), countAlice)

	// Bob marks as read
	err = dto.Message.MarkAsRead(chat.ID, bob.ID)
	require.NoError(t, err)

	// Unread count should be 0 now
	count, err = dto.Message.CountUnread(chat.ID, bob.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDTOMessage_GetUnreadCounts(t *testing.T) {
	resetDB(t)

	alice, bob, chat1 := CreateFriendPair(t)

	// Create another friend pair for Alice
	charles := CreateUser(t, "charles_"+t.Name())
	chat2, _ := CreateFriendship(t, alice, charles)

	// Bob sends to chat1
	_, err := dto.Message.Create(bob.ID, chat1.ID, "Hi Alice", model.ContentTypeText, nil, nil)
	require.NoError(t, err)

	// Charles sends to chat2
	_, err = dto.Message.Create(charles.ID, chat2.ID, "Hello Alice", model.ContentTypeText, nil, nil)
	require.NoError(t, err)

	// Get all unread counts for Alice
	counts, err := dto.Message.GetUnreadCounts(alice.ID)
	require.NoError(t, err)
	assert.Len(t, counts, 2)

	for _, c := range counts {
		if c.ChatID == chat1.ID {
			assert.Equal(t, int64(1), c.Count)
		} else if c.ChatID == chat2.ID {
			assert.Equal(t, int64(1), c.Count)
		}
	}
}

func TestDTOMessage_GetByID_NotFound(t *testing.T) {
	resetDB(t)

	msg, err := dto.Message.GetByID(99999)
	require.NoError(t, err)
	assert.Nil(t, msg)
}

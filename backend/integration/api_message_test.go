//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/model"
)

func TestAPIMessage_SendAndList(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Alice adds Bob → Bob accepts → create friendship + chat
	body := fmt.Sprintf(`{"friend_id":%d}`, bob.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	require.Equal(t, http.StatusCreated, w.Code)

	var fr model.FriendRequest
	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	json.Unmarshal(data, &fr)

	w2 := doRequest(t, engine, http.MethodPut, fmt.Sprintf("/api/friends/%d/accept", fr.ID),
		nil, authHeaders(bob.Token))
	require.Equal(t, http.StatusOK, w2.Code)

	// Get chat ID from alice's friends list
	w3 := doRequest(t, engine, http.MethodGet, "/api/friends", nil, authHeaders(alice.Token))
	require.Equal(t, http.StatusOK, w3.Code)
	resp3 := apiResult(t, w3.Body.Bytes())
	data3, _ := json.Marshal(resp3.Data)
	var friends []model.FriendshipInfo
	json.Unmarshal(data3, &friends)
	require.Len(t, friends, 1)
	chatID := friends[0].ChatID

	// Alice sends a message
	msgBody := fmt.Sprintf(`{"chat_id":%d,"content":"Hello Bob!"}`, chatID)
	w4 := doRequest(t, engine, http.MethodPost, "/api/messages",
		strings.NewReader(msgBody), authHeaders(alice.Token))
	assert.Equal(t, http.StatusCreated, w4.Code)

	resp4 := apiResult(t, w4.Body.Bytes())
	assert.NotNil(t, resp4.Data)
	data4, _ := json.Marshal(resp4.Data)
	var msg model.MessageResponse
	json.Unmarshal(data4, &msg)
	assert.NotZero(t, msg.ID)
	assert.Equal(t, "Hello Bob!", msg.Content)

	// List messages
	w5 := doRequest(t, engine, http.MethodGet,
		fmt.Sprintf("/api/messages?chat_id=%d&offset=0&limit=10", chatID),
		nil, authHeaders(alice.Token))
	assert.Equal(t, http.StatusOK, w5.Code)

	resp5 := apiResult(t, w5.Body.Bytes())
	data5, _ := json.Marshal(resp5.Data)
	var msgs []model.MessageResponse
	json.Unmarshal(data5, &msgs)
	require.Len(t, msgs, 1)
	assert.Equal(t, "Hello Bob!", msgs[0].Content)
}

func TestAPIMessage_SendWithImageMetadata(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Setup friendship
	SetupFriendAndChat(t, engine, alice, bob, func(chatID int64) {
		// Send an image message
		msgBody := fmt.Sprintf(`{
			"chat_id":%d,
			"content":"Check this photo",
			"content_type":"image",
			"file_metadata":{
				"url":"https://oss.example.com/photo.jpg",
				"original_name":"photo.jpg",
				"file_size":204800,
				"mime_type":"image/jpeg",
				"width":800,
				"height":600
			}
		}`, chatID)

		w := doRequest(t, engine, http.MethodPost, "/api/messages",
			strings.NewReader(msgBody), authHeaders(alice.Token))
		assert.Equal(t, http.StatusCreated, w.Code)

		resp := apiResult(t, w.Body.Bytes())
		data, _ := json.Marshal(resp.Data)
		var msg model.MessageResponse
		json.Unmarshal(data, &msg)
		require.NotNil(t, msg.FileMetadata)
		assert.Equal(t, "photo.jpg", msg.FileMetadata.OriginalName)
		assert.Equal(t, 800, msg.FileMetadata.Width)
		assert.Equal(t, 600, msg.FileMetadata.Height)
	})
}

func TestAPIMessage_Send_NotMember(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")
	charles := registerUser(t, engine, "charles_"+t.Name(), "pass123", "Charles")

	// Alice + Bob become friends
	SetupFriendAndChat(t, engine, alice, bob, func(chatID int64) {
		// Charles tries to send to their chat
		msgBody := fmt.Sprintf(`{"chat_id":%d,"content":"Hello"}`, chatID)
		w := doRequest(t, engine, http.MethodPost, "/api/messages",
			strings.NewReader(msgBody), authHeaders(charles.Token))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAPIMessage_MarkAsRead(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Setup friendship
	SetupFriendAndChat(t, engine, alice, bob, func(chatID int64) {
		// Alice sends 2 messages
		for i := 0; i < 2; i++ {
			msgBody := fmt.Sprintf(`{"chat_id":%d,"content":"Msg %d"}`, chatID, i)
			w := doRequest(t, engine, http.MethodPost, "/api/messages",
				strings.NewReader(msgBody), authHeaders(alice.Token))
			require.Equal(t, http.StatusCreated, w.Code)
		}

		// Bob checks unread
		w2 := doRequest(t, engine, http.MethodGet, "/api/messages/unread",
			nil, authHeaders(bob.Token))
		assert.Equal(t, http.StatusOK, w2.Code)
		resp2 := apiResult(t, w2.Body.Bytes())
		data2, _ := json.Marshal(resp2.Data)
		var unread []struct {
			ChatID int64 `json:"chat_id"`
			Count  int64 `json:"count"`
		}
		json.Unmarshal(data2, &unread)
		require.Len(t, unread, 1)
		assert.Equal(t, int64(2), unread[0].Count)

		// Bob marks as read
		readBody := fmt.Sprintf(`{"chat_id":%d}`, chatID)
		w3 := doRequest(t, engine, http.MethodPut, "/api/messages/read",
			strings.NewReader(readBody), authHeaders(bob.Token))
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}

func TestAPIMessage_Unauthorized(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()

	w := doRequest(t, engine, http.MethodGet, "/api/messages?chat_id=1&offset=0&limit=10", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w2 := doRequest(t, engine, http.MethodPost, "/api/messages",
		strings.NewReader(`{"chat_id":1,"content":"hi"}`), nil)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

// SetupFriendAndChat creates friendship between two users and executes callback with chatID.
func SetupFriendAndChat(t *testing.T, engine *gin.Engine, alice, bob *model.AuthResponse, fn func(chatID int64)) {
	t.Helper()

	// Alice adds Bob
	body := fmt.Sprintf(`{"friend_id":%d}`, bob.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	require.Equal(t, http.StatusCreated, w.Code)

	var fr model.FriendRequest
	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	json.Unmarshal(data, &fr)

	// Bob accepts
	w2 := doRequest(t, engine, http.MethodPut, fmt.Sprintf("/api/friends/%d/accept", fr.ID),
		nil, authHeaders(bob.Token))
	require.Equal(t, http.StatusOK, w2.Code)

	// Get chat ID
	w3 := doRequest(t, engine, http.MethodGet, "/api/friends", nil, authHeaders(alice.Token))
	require.Equal(t, http.StatusOK, w3.Code)
	resp3 := apiResult(t, w3.Body.Bytes())
	data3, _ := json.Marshal(resp3.Data)
	var friends []model.FriendshipInfo
	json.Unmarshal(data3, &friends)
	require.Len(t, friends, 1)

	fn(friends[0].ChatID)
}

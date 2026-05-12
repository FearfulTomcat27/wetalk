//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/model"
)

func TestAPIFriend_AddAndAccept(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Alice sends friend request to Bob
	body := fmt.Sprintf(`{"friend_id":%d}`, bob.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	assert.Equal(t, http.StatusCreated, w.Code)

	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	var fr model.FriendRequest
	json.Unmarshal(data, &fr)
	assert.Equal(t, model.FriendStatusPending, fr.Status)

	// Bob should see pending request
	w2 := doRequest(t, engine, http.MethodGet, "/api/friends/pending", nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusOK, w2.Code)
	resp2 := apiResult(t, w2.Body.Bytes())
	data2, _ := json.Marshal(resp2.Data)
	var pending []model.PendingRequest
	json.Unmarshal(data2, &pending)
	require.Len(t, pending, 1)
	assert.Equal(t, alice.User.ID, pending[0].User.ID)

	// Bob accepts
	w3 := doRequest(t, engine, http.MethodPut, fmt.Sprintf("/api/friends/%d/accept", fr.ID),
		nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusOK, w3.Code)

	// Verify friendship exists in both directions
	w4 := doRequest(t, engine, http.MethodGet, "/api/friends", nil, authHeaders(alice.Token))
	assert.Equal(t, http.StatusOK, w4.Code)
	resp4 := apiResult(t, w4.Body.Bytes())
	data4, _ := json.Marshal(resp4.Data)
	var aliceFriends []model.FriendshipInfo
	json.Unmarshal(data4, &aliceFriends)
	require.Len(t, aliceFriends, 1)
	assert.Equal(t, bob.User.ID, aliceFriends[0].FriendID)

	w5 := doRequest(t, engine, http.MethodGet, "/api/friends", nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusOK, w5.Code)
	resp5 := apiResult(t, w5.Body.Bytes())
	data5, _ := json.Marshal(resp5.Data)
	var bobFriends []model.FriendshipInfo
	json.Unmarshal(data5, &bobFriends)
	require.Len(t, bobFriends, 1)
	assert.Equal(t, alice.User.ID, bobFriends[0].FriendID)
}

func TestAPIFriend_AddFriend_Duplicate(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Alice adds Bob
	body := fmt.Sprintf(`{"friend_id":%d}`, bob.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	assert.Equal(t, http.StatusCreated, w.Code)

	// Bob accepts
	var fr model.FriendRequest
	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	json.Unmarshal(data, &fr)

	w2 := doRequest(t, engine, http.MethodPut, fmt.Sprintf("/api/friends/%d/accept", fr.ID),
		nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusOK, w2.Code)

	// Alice tries again — should fail (already friends)
	w3 := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	assert.Equal(t, http.StatusConflict, w3.Code)
}

func TestAPIFriend_AddFriend_Self(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")

	body := fmt.Sprintf(`{"friend_id":%d}`, alice.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIFriend_AddFriend_Unauthorized(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	body := `{"friend_id":2}`
	w := doRequest(t, engine, http.MethodPost, "/api/friends", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIFriend_DeleteFriend(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	alice := registerUser(t, engine, "alice_"+t.Name(), "pass123", "Alice")
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	// Alice sends request
	body := fmt.Sprintf(`{"friend_id":%d}`, bob.User.ID)
	w := doRequest(t, engine, http.MethodPost, "/api/friends",
		strings.NewReader(body), authHeaders(alice.Token))
	assert.Equal(t, http.StatusCreated, w.Code)

	var fr model.FriendRequest
	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	json.Unmarshal(data, &fr)

	// Bob accepts
	w2 := doRequest(t, engine, http.MethodPut, fmt.Sprintf("/api/friends/%d/accept", fr.ID),
		nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusOK, w2.Code)

	// Alice deletes friendship
	w3 := doRequest(t, engine, http.MethodDelete, fmt.Sprintf("/api/friends/%d", fr.ID),
		nil, authHeaders(alice.Token))
	assert.Equal(t, http.StatusOK, w3.Code)

	// Verify both friends lists empty
	w4 := doRequest(t, engine, http.MethodGet, "/api/friends", nil, authHeaders(alice.Token))
	assert.Equal(t, http.StatusOK, w4.Code)
	resp4 := apiResult(t, w4.Body.Bytes())
	data4, _ := json.Marshal(resp4.Data)
	var aliceFriends []model.FriendshipInfo
	json.Unmarshal(data4, &aliceFriends)
	assert.Len(t, aliceFriends, 0)
}

func TestAPIFriend_Accept_NonExistent(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	bob := registerUser(t, engine, "bob_"+t.Name(), "pass123", "Bob")

	w := doRequest(t, engine, http.MethodPut, "/api/friends/99999/accept",
		nil, authHeaders(bob.Token))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
)

type FriendsHandler struct {
	queries *db.Queries
}

func NewFriendsHandler(queries *db.Queries) *FriendsHandler {
	return &FriendsHandler{queries: queries}
}

func (h *FriendsHandler) currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return uuid.Nil, false
	}
	return id, true
}

// SendFriendRequest sends a friend request to another user.
// POST /api/friends/request/:userID
func (h *FriendsHandler) SendFriendRequest(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	targetID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot send friend request to yourself"})
		return
	}

	existing, err := h.queries.GetFriendship(c.Request.Context(), db.GetFriendshipParams{
		RequesterID: currentID,
		AddresseeID: targetID,
	})
	if err == nil {
		if existing.Status == "blocked" {
			c.JSON(http.StatusForbidden, gin.H{"error": "blocked"})
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "friendship already exists"})
		}
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}

	friendship, err := h.queries.SendFriendRequest(c.Request.Context(), db.SendFriendRequestParams{
		RequesterID: currentID,
		AddresseeID: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send request"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// AcceptFriendRequest accepts an incoming friend request.
// POST /api/friends/accept/:friendshipID
func (h *FriendsHandler) AcceptFriendRequest(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	friendshipID, err := uuid.Parse(c.Param("friendshipID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship id"})
		return
	}

	friendship, err := h.queries.GetFriendshipByID(c.Request.Context(), friendshipID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "friendship not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		}
		return
	}

	if friendship.AddresseeID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not the addressee of this request"})
		return
	}

	updated, err := h.queries.UpdateFriendshipStatus(c.Request.Context(), db.UpdateFriendshipStatusParams{
		Status: "accepted",
		ID:     friendshipID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept request"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeclineFriendRequest declines or cancels a friend request.
// POST /api/friends/decline/:friendshipID
func (h *FriendsHandler) DeclineFriendRequest(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	friendshipID, err := uuid.Parse(c.Param("friendshipID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship id"})
		return
	}

	friendship, err := h.queries.GetFriendshipByID(c.Request.Context(), friendshipID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "friendship not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		}
		return
	}

	if friendship.RequesterID != currentID && friendship.AddresseeID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a participant in this friendship"})
		return
	}

	if err := h.queries.DeleteFriendship(c.Request.Context(), friendshipID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decline request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "declined"})
}

// BlockUser blocks another user.
// POST /api/friends/block/:userID
func (h *FriendsHandler) BlockUser(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	targetID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		return
	}

	// Delete any existing friendship in either direction, then insert blocked row.
	existing, err := h.queries.GetFriendship(c.Request.Context(), db.GetFriendshipParams{
		RequesterID: currentID,
		AddresseeID: targetID,
	})
	if err == nil {
		_ = h.queries.DeleteFriendship(c.Request.Context(), existing.ID)
	}

	friendship, err := h.queries.SendFriendRequest(c.Request.Context(), db.SendFriendRequestParams{
		RequesterID: currentID,
		AddresseeID: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	updated, err := h.queries.UpdateFriendshipStatus(c.Request.Context(), db.UpdateFriendshipStatusParams{
		Status: "blocked",
		ID:     friendship.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ListFriends returns accepted friends for the current user.
// GET /api/friends
func (h *FriendsHandler) ListFriends(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	friends, err := h.queries.GetFriends(c.Request.Context(), currentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch friends"})
		return
	}
	if friends == nil {
		friends = []db.GetFriendsRow{}
	}

	c.JSON(http.StatusOK, friends)
}

// ListPendingRequests returns incoming pending friend requests.
// GET /api/friends/requests
func (h *FriendsHandler) ListPendingRequests(c *gin.Context) {
	currentID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	requests, err := h.queries.GetPendingFriendRequests(c.Request.Context(), currentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch requests"})
		return
	}
	if requests == nil {
		requests = []db.GetPendingFriendRequestsRow{}
	}

	c.JSON(http.StatusOK, requests)
}

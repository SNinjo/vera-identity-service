package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.service.GetUsers()
	if err != nil {
		c.Error(err)
		return
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *newUserResponse(&user)
	}

	c.JSON(http.StatusOK, userResponses)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	err := h.service.CreateUser(req.Email)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var uri RequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}
	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	err := h.service.UpdateUser(uri.ID, body.Email)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	var req RequestURI
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}

	err := h.service.DeleteUser(req.ID)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

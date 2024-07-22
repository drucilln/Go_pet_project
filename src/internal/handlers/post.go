package handlers

import (
	"fmt"
	"log"
	"net/http"
	"project/internal/database"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePostPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "post.html", nil)
}

func (h *Handler) CreatePostHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("CreatePostHandler: user_id not found in context")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	var post database.Post
	post.UserID = uint(userID.(uint))
	post.Content = c.PostForm("content")

	log.Println("CreatePostHandler: received data:", post)

	if err := h.db.Create(&post).Error; err != nil {
		log.Println("CreatePostHandler: error creating post:", err)
		c.HTML(http.StatusInternalServerError, "create_post.html", gin.H{
			"Error":   "Failed to create post",
			"Content": post.Content,
		})
		return
	}

	username := c.GetString("username")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/%s", username))

}

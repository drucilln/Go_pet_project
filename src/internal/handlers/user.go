package handlers

import (
	"fmt"
	"net/http"
	"project/internal/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UserPageHandler(c *gin.Context) {
	username := c.Param("username")

	var user database.Users

	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil { // Проверка на существоаание пользователя
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	if c.Request.Method == http.MethodPost { // Поляе для публикации поста
		content := c.PostForm("content")
		if content != "" {
			session := sessions.Default(c)
			currentUserID := session.Get("user_id").(uint)
			post := database.Post{
				UserID:  currentUserID,
				Content: content,
			}
			if err := h.db.Create(&post).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "user.html", gin.H{
					"Error": "Failed to create post",
				})
				return
			}
		}
	}

	var posts []database.Post

	if err := h.db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&posts).Error; err != nil { // Поиск постов пользователя
		c.HTML(http.StatusInternalServerError, "user.html", gin.H{
			"Error": "Failed to load posts",
		})
		return
	}

	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(uint)
	currentUsername := session.Get("username")

	var isFollowing bool
	if currentUserID != 0 {
		var follow database.Follow
		if err := h.db.Where("follower_id = ? AND followed_id = ?", currentUserID, user.ID).First(&follow).Error; err == nil { // Проверка на наличие подписки
			isFollowing = true
		}
	}

	var userFollowersCount, userFollowingsCount int64

	h.db.Model(&database.Follow{}).Where("followed_id = ?", user.ID).Count(&userFollowersCount)
	h.db.Model(&database.Follow{}).Where("follower_id = ?", user.ID).Count(&userFollowingsCount)

	c.HTML(http.StatusOK, "user.html", gin.H{
		"Username":        username,
		"UserID":          user.ID,
		"Posts":           posts,
		"CurrentUserID":   currentUserID,
		"CurrentUsername": currentUsername,
		"IsFollowing":     isFollowing,
		"FollowersCount":  userFollowersCount,
		"FollowingsCount": userFollowingsCount,
	})
}

func (h *Handler) DeletePostHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	postID := c.Param("id")
	var post database.Post

	if err := h.db.Where("id = ? AND user_id = ?", postID, userID).First(&post).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "Post not found",
		})
		return
	}

	if err := h.db.Delete(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "user.html", gin.H{
			"Error": "Failed to delete post",
		})
		return
	}
	username := session.Get("username")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/%s", username))

}

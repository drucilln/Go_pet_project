package handlers

import (
	"fmt"
	"net/http"
	"project/internal/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FollowersPageHandler(c *gin.Context) {
	username := c.Param("username")

	var user database.Users
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	var followers []database.Follow
	if err := h.db.Where("followed_id = ?", user.ID).Find(&followers).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "followers.html", gin.H{
			"Error": "Failed to load followers",
		})
		return
	}
	var follows []struct {
		database.Users
		IsFollowing bool
	}
	session := sessions.Default(c)
	currentUsername := session.Get("username")
	currentUserID := session.Get("user_id").(uint)

	for _, follow := range followers {
		var follower database.Users
		if err := h.db.Where("id = ?", follow.FollowerID).First(&follower).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "followers.html", gin.H{
				"Error": "Failed to load follower user details",
			})
			return
		}
		var isFollowing bool
		if err := h.db.Where("follower_id = ? AND followed_id = ?", currentUserID, follower.ID).First(&database.Follow{}).Error; err == nil {
			isFollowing = true
		}

		follows = append(follows, struct {
			database.Users
			IsFollowing bool
		}{
			Users:       follower,
			IsFollowing: isFollowing,
		})
	}

	c.HTML(http.StatusOK, "followers.html", gin.H{
		"Username":        username,
		"Followers":       follows,
		"CurrentUsername": currentUsername,
	})
}

func (h *Handler) FollowingPageHandler(c *gin.Context) {
	username := c.Param("username")

	var user database.Users
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	var followings []database.Follow
	if err := h.db.Where("follower_id = ?", user.ID).Find(&followings).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "following.html", gin.H{
			"Error": "Failed to load followers",
		})
		return
	}

	var follows []struct {
		database.Users
		IsFollowing bool
	}

	session := sessions.Default(c)
	currentUsername := session.Get("username")
	currentUserID := session.Get("user_id").(uint)

	for _, follow := range followings {
		var following database.Users
		if err := h.db.Where("id = ?", follow.FollowedID).First(&following).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "following.html", gin.H{
				"Error": "Failed to load following user details",
			})
			return
		}

		var isFollowing bool
		if err := h.db.Where("follower_id = ? AND followed_id = ?", currentUserID, following.ID).First(&database.Follow{}).Error; err == nil {
			isFollowing = true
		}
		follows = append(follows, struct {
			database.Users
			IsFollowing bool
		}{
			Users:       following,
			IsFollowing: isFollowing,
		})
	}

	c.HTML(http.StatusOK, "following.html", gin.H{
		"Username":        username,
		"Followings":      follows,
		"CurrentUsername": currentUsername,
	})
}

func (h *Handler) FollowHandler(c *gin.Context) {
	session := sessions.Default(c)
	followerID := session.Get("user_id")
	followedUsername := c.Param("username")

	var user database.Users
	if err := h.db.Where("username = ?", followedUsername).First(&user).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	follow := database.Follow{
		FollowerID: followerID.(uint),
		FollowedID: user.ID,
	}
	if err := h.db.Create(&follow).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "user.html", gin.H{
			"Error": "Failed to follow user",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/%s", followedUsername))
}

func (h *Handler) UnfollowHandler(c *gin.Context) {
	session := sessions.Default(c)
	followerID := session.Get("user_id")
	followedUsername := session.Get("username")

	var followedUser database.Users
	if err := h.db.Where("username = ?", followedUsername).First(&followedUser).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	var follow database.Follow
	if err := h.db.Where("follower_id = ? AND followed_id + ?", followerID, followedUser.ID).First(&follow).Error; err != nil {
		c.HTML(http.StatusNotFound, "user.html", gin.H{
			"Error": "Follow relationship not found",
		})
		return
	}

	if err := h.db.Delete(&follow).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "user.html", gin.H{
			"Error": "Failed to unfollow user",
		})
		return
	}

	c.Redirect(http.StatusOK, fmt.Sprintf("/user/%s", followedUsername))
}

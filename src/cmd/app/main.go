package main

import (
	db "project/internal/database"
	"project/internal/handlers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	db := db.InitDB()

	r := gin.Default()

	h := handlers.NewHandler(db)

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("web/templates/*")

	r.GET("/login", h.LoginPageHandler)
	r.POST("/login", h.LoginHandler)
	r.GET("/logout", h.LogoutHandler)
	r.GET("/register", h.RegisterPageHandler)
	r.POST("/register", h.RegisterHandler)
	r.GET("/user/:username", h.UserPageHandler)
	r.GET("/user/:username/followers", h.FollowersPageHandler)
	r.GET("/user/:username/followings", h.FollowingPageHandler)
	r.GET("/post/create", h.AuthRequired(), h.CreatePostPageHandler)
	r.POST("/post/create", h.AuthRequired(), h.CreatePostHandler)
	r.POST("/post/delete/:id", h.AuthRequired(), h.DeletePostHandler)
	r.POST("/user/follow/:username", h.AuthRequired(), h.FollowHandler)
	r.POST("/user/unfollow/:username", h.AuthRequired(), h.UnfollowHandler)
	r.GET("/message/:username", h.AuthRequired(), h.MessagePageHandler)
	r.GET("/ws/messages/:username", h.AuthRequired(), h.MessageWebSocketHandler)
	r.POST("/message/send/:username", h.AuthRequired(), h.SendMessageHandler)
	r.Run(":8080")

}

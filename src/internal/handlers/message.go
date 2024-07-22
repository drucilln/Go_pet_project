package handlers

import (
	"fmt"
	"log"
	"net/http"
	"project/internal/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) MessageWebSocketHandler(c *gin.Context) {
	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(uint)
	currentUsername := session.Get("username").(string)
	receiverUsername := c.Param("username")

	var receiver database.Users
	if err := h.db.Where("username = ?", receiverUsername).First(&receiver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Error": "User not found",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: ", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	h.userConnection[currentUsername] = conn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.userConnection, currentUsername)
		h.mu.Unlock()
	}()

	for {
		var msg database.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json.", err)
			break
		}
		msg.SenderID = currentUserID
		msg.ReceiverID = receiver.ID

		if err := h.db.Create(&msg).Error; err != nil {
			log.Println("Failed to save message: ", err)
			continue
		}

		h.mu.Lock()
		if receiverConn, ok := h.userConnection[receiverUsername]; ok {
			receiverConn.WriteJSON(msg)
		}
		if senderConn, ok := h.userConnection[currentUsername]; ok {
			senderConn.WriteJSON(msg)
		}
		h.mu.Unlock()
	}
}

func (h *Handler) MessagePageHandler(c *gin.Context) {
	receiverUsername := c.Param("username")

	var user database.Users
	if err := h.db.Where("username = ?", receiverUsername).First(&user).Error; err != nil {
		c.HTML(http.StatusNotFound, "message.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(uint)
	currentUsername := session.Get("username").(string)

	var messages []database.Message
	if err := h.db.Where("sender_id = ? AND receiver_id = ?", user.ID, currentUserID).Or("sender_id = ? AND receiver_id = ?", currentUserID, user.ID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "message.html", gin.H{
			"Error": "Failed to load messages",
		})
	}

	c.HTML(http.StatusOK, "message.html", gin.H{
		"Messages":         messages,
		"ReceiverUsername": receiverUsername,
		"CurrentUsername":  currentUsername,
	})
}

func (h *Handler) SendMessageHandler(c *gin.Context) {
	receiverUsername := c.Param("username")

	var user database.Users
	if err := h.db.Where("username = ?", receiverUsername).First(&user).Error; err != nil {
		c.HTML(http.StatusNotFound, "message.html", gin.H{
			"Error": "User not found",
		})
		return
	}

	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(uint)

	var message database.Message
	message.Content = c.PostForm("content")
	message.ReceiverID = user.ID
	message.SenderID = currentUserID
	if err := h.db.Create(&message).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "message.html", gin.H{
			"Error":   "Failed to send message",
			"Content": message.Content,
		})
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/message/%s", receiverUsername))
}

func (h *Handler) SendMessage(senderUsername, receiverUsername, content string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if receiverConn, ok := h.userConnection[receiverUsername]; ok {
		msg := database.Message{
			Content: content,
		}
		receiverConn.WriteJSON(msg)
	}
}

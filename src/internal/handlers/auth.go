package handlers

import (
	"fmt"
	"log"
	"net/http"
	"project/internal/config"
	"project/internal/database"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	db             *gorm.DB
	userConnection map[string]*websocket.Conn
	mu             sync.Mutex
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:             db,
		userConnection: make(map[string]*websocket.Conn),
	}
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString, err := c.Cookie("jwt")
		if err != nil {
			// c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			// 	"Error": "Authorization header is required",
			// })
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected singing method: %v", token.Header["alg"])
			}
			return config.JWTSecret, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"Error": err.Error(),
			})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := uint(claims["sub"].(float64))
			c.Set("user_id", userID)
			c.Set("username", claims["username"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Next()

	}
}

func (h *Handler) LogoutHandler(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "localhost", false, true)

	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.Redirect(http.StatusSeeOther, "/login")
}

func (h *Handler) RegisterHandler(c *gin.Context) {
	var user database.Users

	user.Username = c.PostForm("username")
	user.Password = c.PostForm("password")

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Println("RegisterHandler: failed to hash password:", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"Error":    "Failed to hash password",
			"Username": user.Username,
		})
		return
	}
	user.Password = hashedPassword

	err = h.db.Create(&user).Error
	if err != nil {
		log.Println("RegisterHandler: error creating user:", err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			log.Println("RegisterHandler: username already exists")
			c.HTML(http.StatusConflict, "register.html", gin.H{
				"Error":    "Username already exists",
				"Username": user.Username,
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"Error":    err.Error(),
			"Username": user.Username,
		})
		return
	}

	token, err := createJWTToken(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Error": "Failed to create token",
		})
		return
	}

	c.SetCookie("jwt", token, 3600, "/", "localhost", false, true)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Save()

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/%s", user.Username))

	// c.Redirect(http.StatusSeeOther, "/login")
}

func (h *Handler) RegisterPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordhash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (h *Handler) LoginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user database.Users

	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error":    "Invalid username or password",
			"Username": username,
		})
		return
	}

	if !checkPasswordhash(password, user.Password) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error":    "Invalid username or password",
			"Username": username,
		})
		return
	}

	token, err := createJWTToken(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Error": "Failed to create token",
		})
		return
	}

	c.SetCookie("jwt", token, 3600, "/", "localhost", false, true)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Save()

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/%s", username))
}

func (h *Handler) LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func createJWTToken(user database.Users) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

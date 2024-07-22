package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Username string    `gorm:"unique;not null"`
	Password string    `gorm:"not null"`
	Posts    []Post    `gorm:"foreignKey:UserID"`
	Messages []Message `gorm:"foreignKey:SenderID"`
}

type Follow struct {
	gorm.Model
	FollowerID uint
	FollowedID uint
	Follower   Users `gorm:"foreignKey:FollowerID"`
	Followed   Users `gorm:"foreignKey:FollowedID"`
}

type Post struct {
	gorm.Model
	UserID  uint   `gorm:"not null"`
	Content string `gorm:"not null"`
	User    Users  `gorm:"foreignKey:UserID"`
}

type Message struct {
	gorm.Model
	SenderID   uint   `gorm:"not null"`
	ReceiverID uint   `gorm:"not null"`
	Content    string `gorm:"not null"`
	// Sender     Users  `gorm:"foreignKey:SenderID"`
	// Receiver   Users  `gorm:"foreignKey:ReceiverID"`
}

func InitDB() *gorm.DB {
	dsn := "host=localhost user=polaykov.art dbname=blog-project sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	// db.Migrator().DropTable(&Users{}, &Follow{}, &Post{}, &Message{})
	err = db.AutoMigrate(&Users{}, &Follow{}, &Post{}, &Message{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}
	return db
}

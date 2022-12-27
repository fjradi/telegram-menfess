package adapter

import (
	"database/sql"
	"log"
	"telegram/internal/domain"
)

// Postgres is a struct that holds the database connection
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a new Postgres struct
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

// SaveMessage saves a message to the database
func (p *Postgres) SaveMessage(message domain.Message) error {
	var fileId *string
	if len(message.Photo) > 0 {
		fileId = &message.Photo[0].FileId
	}

	var text string
	if message.Text != "" {
		text = message.Text
	} else if message.Caption != "" {
		text = message.Caption
	}

	_, err := p.db.Exec("INSERT INTO message (message_id, chat_id, text, username, file_id) VALUES ($1, $2, $3, $4, $5)", message.Id, message.Chat.Id, text, message.From.Username, fileId)
	if err != nil {
		log.Printf("error inserting message into database %s", err.Error())
		return err
	}

	return nil
}

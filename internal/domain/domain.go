package domain

import (
	"encoding/json"
	"errors"
	"log"
)

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Id      int     `json:"message_id"`
	Text    string  `json:"text"`
	Chat    Chat    `json:"chat"`
	From    User    `json:"from"`
	Photo   []Photo `json:"photo"`
	Caption string  `json:"caption"`
}

type Chat struct {
	Id int `json:"id"`
}

type User struct {
	Id       int     `json:"id"`
	Username *string `json:"username"`
}

type Photo struct {
	FileId string `json:"file_id"`
}

func (m Message) ToJson() ([]byte, error) {
	jsonByte, err := json.Marshal(m)
	if err != nil {
		log.Printf("error marshalling message %s", err.Error())
		return nil, err
	}

	return jsonByte, nil
}

type Gender int

const (
	Male   Gender = iota
	Female Gender = iota
)

var InvalidMessageError = errors.New("invalid message")

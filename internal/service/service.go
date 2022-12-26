package service

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"telegram/internal/domain"
	"telegram/internal/port"
)

type Service struct {
	repo                 port.Repository
	outGoingMessageQueue port.OutgoingMessageQueue
	inGoingMessageQueue  port.IngoingMessageQueue
	botToken             string
	channelChatId        int
}

func NewService(repo port.Repository, outGoingMessageQueue port.OutgoingMessageQueue, inGoingMessageQueue port.IngoingMessageQueue, botToken string, channelChatId int) *Service {
	svc := &Service{repo: repo, outGoingMessageQueue: outGoingMessageQueue, inGoingMessageQueue: inGoingMessageQueue, botToken: botToken, channelChatId: channelChatId}

	go func() {
		err := svc.start()
		if err != nil {
			log.Panic(err)
		}
	}()

	return svc
}

func (s *Service) start() error {
	messages, err := s.inGoingMessageQueue.StreamMessages()
	if err != nil {
		return err
	}

	for message := range messages {
		go func(message domain.Message) {
			s.repo.SaveMessage(message)
		}(message)
	}

	return nil
}

func (s *Service) ReceiveMessage(message domain.Message) error {
	if message.Chat.Id == s.channelChatId {
		return nil
	}

	errChan := make(chan error)

	go func() {
		err := s.sendMessage(message)
		errChan <- err
	}()

	go func() {
		err := s.outGoingMessageQueue.PublishMessage(message)
		errChan <- err
	}()

	var errs []error
	for i := 0; i < 2; i++ {
		err := <-errChan
		errs = append(errs, err)
	}

	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) sendMessage(message domain.Message) error {
	sendMessageEndpoint := "https://api.telegram.org/bot" + s.botToken

	var text string
	if message.Text != "" {
		text = message.Text
	} else if message.Caption != "" {
		text = message.Caption
	}

	if message.From.Username != nil {
		text = text + " @" + *message.From.Username
	}

	var fileId *string
	if len(message.Photo) > 0 {
		fileId = &message.Photo[0].FileId
	}

	var response *http.Response
	var err error
	if fileId != nil {
		response, err = http.PostForm(sendMessageEndpoint+"/sendPhoto", url.Values{
			"chat_id": {strconv.Itoa(s.channelChatId)},
			"photo":   {*fileId},
			"caption": {text},
		})
	} else {
		response, err = http.PostForm(sendMessageEndpoint+"/sendMessage", url.Values{
			"chat_id": {strconv.Itoa(s.channelChatId)},
			"text":    {text},
		})
	}

	if err != nil {
		log.Printf("error sending message %s", err.Error())
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		var bodyBytes, errRead = ioutil.ReadAll(response.Body)
		if errRead != nil {
			log.Printf("error in parsing telegram answer %s", errRead.Error())
			return errRead
		}

		bodyString := string(bodyBytes)
		log.Printf("Body of Telegram Response: %s", bodyString)

		return errors.New("telegram response status code is not 200")
	}

	return nil
}

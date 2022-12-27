package service

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"telegram/internal/domain"
	"telegram/internal/port"
	"time"
)

type Service struct {
	repo                 port.Repository
	memoryRepo           port.MemoryRepository
	outGoingMessageQueue port.OutgoingMessageQueue
	inGoingMessageQueue  port.IngoingMessageQueue
	botToken             string
	channelChatId        int
}

func NewService(repo port.Repository, memoryRepo port.MemoryRepository, outGoingMessageQueue port.OutgoingMessageQueue, inGoingMessageQueue port.IngoingMessageQueue, botToken string, channelChatId int) *Service {
	svc := &Service{repo: repo, memoryRepo: memoryRepo, outGoingMessageQueue: outGoingMessageQueue, inGoingMessageQueue: inGoingMessageQueue, botToken: botToken, channelChatId: channelChatId}

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

	latestSentMessageTimestamp, err := s.getLatestSentMessageTimestamp(message.From.Id)
	if err != nil {
		return err
	}

	if time.Now().Sub(latestSentMessageTimestamp) >= 5*time.Minute {
		err = s.forwardMessage(message)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.memoryRepo.SaveLatestSentMessageTimestamp(message.From.Id, time.Now().Unix())
			s.replyMessage(message.Chat.Id, "Pesan kamu telah terkirim ke channel")
		}()
		go func() {
			defer wg.Done()
			s.outGoingMessageQueue.PublishMessage(message)
		}()

		wg.Wait()
		return nil
	}

	ellapsedTime := time.Now().Sub(latestSentMessageTimestamp)
	remainingDuration := 5*time.Minute - ellapsedTime
	remainingDurationMinute := int(remainingDuration.Minutes())
	remainingDurationSecond := int(remainingDuration.Seconds()) - remainingDurationMinute*60

	err = s.replyMessage(message.Chat.Id, "Mohon tunggu "+strconv.Itoa(remainingDurationMinute)+" menit "+strconv.Itoa(remainingDurationSecond)+" detik lagi untuk mengirim pesan")
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getLatestSentMessageTimestamp(userId int) (time.Time, error) {
	latestSentMessageTimestamp, err := s.memoryRepo.GetLatestSentMessageTimestamp(userId)
	if err != nil {
		return time.Time{}, err
	}

	if latestSentMessageTimestamp == 0 {
		return time.Time{}, nil
	}

	return time.Unix(latestSentMessageTimestamp, 0), nil
}

func (s *Service) forwardMessage(message domain.Message) error {
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

	var body url.Values
	if len(message.Photo) > 0 {
		sendMessageEndpoint = sendMessageEndpoint + "/sendPhoto"
		fileId := message.Photo[0].FileId
		body = url.Values{
			"chat_id": {strconv.Itoa(s.channelChatId)},
			"photo":   {fileId},
			"caption": {text},
		}
	} else {
		sendMessageEndpoint = sendMessageEndpoint + "/sendMessage"
		body = url.Values{
			"chat_id": {strconv.Itoa(s.channelChatId)},
			"text":    {text},
		}
	}

	response, err := http.PostForm(sendMessageEndpoint, body)

	if err != nil {
		log.Printf("error sending message %s", err.Error())
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logResponse(response)
	}

	return nil
}

func (s *Service) replyMessage(chatId int, replyText string) error {
	replyMessageEndpoint := "https://api.telegram.org/bot" + s.botToken + "/sendMessage"

	response, err := http.PostForm(replyMessageEndpoint, url.Values{
		"chat_id": {strconv.Itoa(chatId)},
		"text":    {replyText},
	})

	if err != nil {
		log.Printf("error sending message %s", err.Error())
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logResponse(response)
	}

	return nil
}

func (s *Service) logResponse(response *http.Response) error {
	bodyBytes, errRead := ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
		return errRead
	}

	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)

	return errors.New("telegram response status code is not 200")
}

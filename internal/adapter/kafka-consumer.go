package adapter

import (
	"encoding/json"
	"log"
	"telegram/internal/domain"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	topic    string
}

func NewKafkaConsumer(consumer *kafka.Consumer) (*KafkaConsumer, error) {
	k := &KafkaConsumer{consumer: consumer, topic: "telegram"}
	err := k.consumer.Subscribe(k.topic, nil)
	if err != nil {
		log.Printf("error subscribing to topic %s", err.Error())
		return nil, err
	}

	return k, nil
}

func (k *KafkaConsumer) StreamMessages() (<-chan domain.Message, error) {
	messages := make(chan domain.Message)

	go func() {
		for {
			msg, err := k.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("error reading message %s", err.Error())
			} else {
				var message domain.Message
				err = json.Unmarshal(msg.Value, &message)
				if err != nil {
					log.Printf("error unmarshalling message %s", err.Error())
				} else {
					messages <- message
				}
			}
		}
	}()

	return messages, nil
}

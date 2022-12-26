package adapter

import (
	"log"
	"strconv"
	"telegram/internal/domain"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(producer *kafka.Producer) *KafkaProducer {
	return &KafkaProducer{producer: producer, topic: "telegram"}
}

func (k *KafkaProducer) PublishMessage(message domain.Message) error {
	messageJson, err := message.ToJson()
	if err != nil {
		return err
	}

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          messageJson,
		Key:            []byte(strconv.Itoa(message.Chat.Id)),
	}, nil)

	if err != nil {
		log.Printf("error producing message %s", err.Error())
		return err
	}

	return nil
}

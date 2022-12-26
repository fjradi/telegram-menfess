package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"telegram/internal/adapter"
	"telegram/internal/service"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_BOOTSTRAP_SERVERS")})
	if err != nil {
		log.Panicf("Failed to create producer: %s", err)
	}
	log.Println("Producer created")
	defer producer.Close()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Panicf("Failed to create consumer: %s", err)
	}
	log.Println("Consumer created")
	defer consumer.Close()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Panicf("Failed to connect to database: %s", err)
	}
	log.Println("Connected to database")
	defer db.Close()

	kafkaProducer := adapter.NewKafkaProducer(producer)
	kafkaConsumer, err := adapter.NewKafkaConsumer(consumer)
	if err != nil {
		log.Panicf("Failed to create kafka consumer: %s", err)
	}
	postgres := adapter.NewPostgres(db)

	channelChatId := os.Getenv("CHANNEL_CHAT_ID")
	channelChatIdInt, err := strconv.Atoi(channelChatId)
	if err != nil {
		log.Panicf("Failed to convert channel chat id to int: %s", err)
	}

	myService := service.NewService(postgres, kafkaProducer, kafkaConsumer, os.Getenv("TELEGRAM_BOT_TOKEN"), channelChatIdInt)

	restHandler := adapter.NewRestHandler(myService)

	http.HandleFunc("/", restHandler.Handle)
	log.Printf("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"os"
	"strconv"
	"telegram/internal/adapter"
	"telegram/internal/service"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

//go:embed cert.pem
var certFile []byte

//go:embed key.key
var keyFile []byte

func main() {
	godotenv.Load()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     os.Getenv("KAFKA_USERNAME"),
		"sasl.password":     os.Getenv("KAFKA_PASSWORD"),
	})
	if err != nil {
		log.Panicf("Failed to create producer: %s", err)
	}
	log.Println("Producer created")
	defer producer.Close()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"group.id":           "myGroup",
		"auto.offset.reset":  "earliest",
		"security.protocol":  "SASL_SSL",
		"sasl.mechanisms":    "PLAIN",
		"sasl.username":      os.Getenv("KAFKA_USERNAME"),
		"sasl.password":      os.Getenv("KAFKA_PASSWORD"),
		"session.timeout.ms": 45000,
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		DB: 0,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Panicf("Failed to connect to redis: %s", err)
	}
	log.Println("Connected to redis")
	defer redisClient.Close()

	kafkaProducer := adapter.NewKafkaProducer(producer)
	kafkaConsumer, err := adapter.NewKafkaConsumer(consumer)
	if err != nil {
		log.Panicf("Failed to create kafka consumer: %s", err)
	}
	postgres := adapter.NewPostgres(db)
	redis := adapter.NewRedis(redisClient)

	channelChatId := os.Getenv("CHANNEL_CHAT_ID")
	channelChatIdInt, err := strconv.Atoi(channelChatId)
	if err != nil {
		log.Panicf("Failed to convert channel chat id to int: %s", err)
	}

	myService := service.NewService(postgres, redis, kafkaProducer, kafkaConsumer, os.Getenv("TELEGRAM_BOT_TOKEN"), channelChatIdInt)

	restHandler := adapter.NewRestHandler(myService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", restHandler.Handle)
	handler := cors.Default().Handler(mux)

	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		log.Panicf("Failed to load certificate: %s", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	server := &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: config,
	}
	log.Printf("Server is running on port 443")
	if err = server.ListenAndServeTLS("", ""); err != nil {
		log.Panicf("Failed to start server: %s", err)
	} else {
		log.Println("Server stopped")
	}
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/jackc/pgx/v4"
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready       chan bool
	conn        pgx.Conn
	logProducer sarama.AsyncProducer
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	consumer.conn = *conn
	// Mark the consumer as ready
	close(consumer.ready)

	consumer.logProducer = newAccessLogProducer([]string{os.Getenv("KAFKA_URL")})
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	consumer.conn.Close(context.Background())
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		var data ChangeBalanceStruct
		err := json.Unmarshal(message.Value, &data)
		if err != nil {
			log.Printf("Marshall error: %v\n", err)
			session.MarkMessage(message, "")
			continue
		}
		if data.TypeOperation == "withdraw" {
			data.Money = -data.Money
		} else if data.TypeOperation != "deposit" {
			log.Println("Unknown operation type")
			session.MarkMessage(message, "")
			continue
		}
		var balance int
		err = consumer.conn.QueryRow(context.Background(), "SELECT balance FROM bank_accounts WHERE id=$1", data.AccountID).Scan(&balance)
		if err != nil {
			log.Println("SELECT balance is unsuccessfull")
			session.MarkMessage(message, "")
			continue
		}
		balance = balance + data.Money
		if balance < 0 {
			log.Println("Illegal withdraw")
			session.MarkMessage(message, "")
			continue
		}
		_, err = consumer.conn.Exec(context.Background(), "UPDATE bank_accounts SET balance=$1 WHERE id=$2;", balance, data.AccountID)
		if err != nil {
			log.Fatalln("Update failed")
			session.MarkMessage(message, "")
			continue
		}

		consumer.logProducer.Input() <- &sarama.ProducerMessage{
			Topic: "successfull-operations",
			Key:   sarama.StringEncoder(data.TypeOperation),
			Value: &data,
		}
		log.Printf("Consumed message offset %d\n", message.Offset)
		session.MarkMessage(message, "")
	}

	return nil
}

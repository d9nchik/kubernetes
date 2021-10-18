package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	"github.com/jackc/pgx/v4"
)

type ChangeBalanceStruct struct {
	AccountID     int
	Money         int
	TypeOperation string `json:"type"`
}

func main() {
	// KAFKA URL should be in format address:9092
	consumer, err := sarama.NewConsumer([]string{os.Getenv("KAFKA_URL")}, sarama.NewConfig())
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition("change-balance", 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	consumed := 0
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var data ChangeBalanceStruct
			err := json.Unmarshal(msg.Value, &data)
			if err != nil {
				log.Fatalf("Marshall error: %v\n", err)
				continue
			}
			if data.TypeOperation == "withdraw" {
				data.Money = -data.Money
			} else if data.TypeOperation != "deposit" {
				log.Fatalln("Unknown operation type")
				continue
			}

			var balance int
			err = conn.QueryRow(context.Background(), "SELECT balance FROM bank_accounts WHERE id=$1", data.AccountID).Scan(&balance)
			if err != nil {
				log.Fatalln("SELECT balance is unsuccessfull")
				continue
			}
			balance = balance + data.Money
			if balance < 0 {
				log.Fatalln("Illegal withdraw")
				continue
			}
			_, err = conn.Exec(context.Background(), "UPDATE bank_accounts SET balance= WHERE id=$1;", data.AccountID, balance)
			if err != nil {
				log.Fatalln("Update failed")
				continue
			}
			log.Printf("Consumed message offset %d\n", msg.Offset)
			consumed++
		case <-signals:
			break ConsumerLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
}

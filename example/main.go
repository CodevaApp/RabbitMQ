package main

import (
	"fmt"
	"os"

	rmq "github.com/codevaapp/rabbitmq"

	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	serverURL := os.Getenv("AMQP_SERVER_URL")
	options := rmq.RabbitMQClientOptions{
		URL:          serverURL,
		RetryAttempt: 5,
	}

	rabbitMQ, err := rmq.NewRabbitMQClient(options)
	if err != nil {
		return
	}

	queue, err := rabbitMQ.CreateQueue("test2", true, true, false, false, nil)
	if err != nil {
		return
	}

	err = rabbitMQ.CreateExchange("exchange", "fanout", true, true, false, false, nil)
	if err != nil {
		return
	}

	err = rabbitMQ.BindQueueWithExchange(queue.Name, "", "exchange", false, nil)
	if err != nil {
		return
	}

	err = rabbitMQ.CreateMessgeChannel(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return
	}

	for {
		msg, err := rabbitMQ.ConsumeMessageChannel()
		if err != nil {
			return
		}

		fmt.Println(msg)

	}
}

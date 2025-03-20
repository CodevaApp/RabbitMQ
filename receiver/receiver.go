package receiver

import (
	"github.com/codevaapp/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Sender embeds RabbitMQClient and provides methods for sending messages
type Receiver struct {
	rabbitmq.RabbitMQClient // Embed the RabbitMQClient struct
}

func (rc *Receiver) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.Ch.Consume(
		queue,
		consumer,
		autoAck,
		false,
		false,
		false,
		nil,
	)
}

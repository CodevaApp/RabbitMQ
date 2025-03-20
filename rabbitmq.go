package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	// Connection that's used
	Conn *amqp.Connection
	// The channel the processes/sends messages
	Ch *amqp.Channel
}

// Will spawn a new connection
func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, host, vhost))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Connect and return a RabbitMQclient with an open connection
func NewRabbitMQClient(conn *amqp.Connection) (RabbitMQClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return RabbitMQClient{}, err
	}

	return RabbitMQClient{
		Conn: conn,
		Ch:   ch,
	}, nil
}

func (rc *RabbitMQClient) Close() error {
	if err := rc.Ch.Close(); err != nil {
		return err
	}

	return rc.Conn.Close()
}

func (rc *RabbitMQClient) CreateQueue(queueName string, durable, autodelete bool) error {
	_, err := rc.Ch.QueueDeclare(queueName, durable, autodelete, false, false, nil)
	return err
}

func (rc *RabbitMQClient) CreateBinding(name, binding, exchange string) error {
	return rc.Ch.QueueBind(name, binding, exchange, false, nil)
}

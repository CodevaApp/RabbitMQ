package rabbitmq

import (
	"github.com/labstack/gommon/log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	connURL      string
	errCh        <-chan *amqp.Error
	msgCh        <-chan amqp.Delivery
	retryAttempt int
}

type RabbitMQClientOptions struct {
	URL          string
	RetryAttempt int
}

// NewRabbitMQClient creates a new instance of RabbitMQClient using the provided options.
// It initializes the connection URL and retry attempt count, then attempts to establish
// a connection to the RabbitMQ server. If the connection fails, an error is returned.
func NewRabbitMQClient(options RabbitMQClientOptions) (*RabbitMQClient, error) {
	rabbitmq := &RabbitMQClient{
		connURL:      options.URL,
		retryAttempt: options.RetryAttempt,
	}

	if err := rabbitmq.connect(); err != nil {
		return nil, err
	}

	return rabbitmq, nil
}

// connect tries to connect to RabbitMQ server up to retryAttempt times if
// a connection error occurs.
func (rc *RabbitMQClient) connect() error {
	var err error

	rc.conn, err = amqp.Dial(rc.connURL)
	if err != nil {
		log.Error("Error creating connection", err.Error())
		return err
	}

	rc.ch, err = rc.conn.Channel()
	if err != nil {
		log.Error("Error creating a channel", err.Error())
		return err
	}

	rc.errCh = rc.conn.NotifyClose(make(chan *amqp.Error))

	return nil
}

// reconnect tries to reconnect to RabbitMQ server up to retryAttempt times if
// a connection error occurs.
func (rc *RabbitMQClient) reconnect() {
	attempt := rc.retryAttempt

	for attempt != 0 {
		log.Info("RabbitMQ reconnecting")
		if err := rc.connect(); err != nil {
			attempt--
			log.Error("RabbitMQ retry conntection error", err.Error())
		}
	}
}

// Close closes the channel and connection to the RabbitMQ server.
// If there is an error closing the channel or connection, the error is returned.
func (rc *RabbitMQClient) Close() error {
	if err := rc.ch.Close(); err != nil {
		return err
	}

	if err := rc.conn.Close(); err != nil {
		return err
	}

	rc.ch.Close()
	rc.conn.Close()

	return nil
}

// ConsumeMessageChannel consumes a message from a RabbitMQ queue and returns the
// message body as a byte array. If the connection to RabbitMQ is closed, the
// function will reconnect and try consuming again. If there is an error consuming
// the message, the function will return the error.
func (rc *RabbitMQClient) ConsumeMessageChannel() (jsonBytes []byte, err error) {
	select {
	case err := <-rc.errCh:
		log.Warn("RabbitMQ has error from notifyCloseChannel", err.Error())
		rc.reconnect()
	case msg := <-rc.msgCh:
		return msg.Body, nil
	}

	return nil, nil
}

// CreateQueue declares a RabbitMQ queue. If the queue does not exist, it is
// created. If the queue already exists, the queue's properties are checked,
// but any mismatched properties do not cause an exception.
//
// The queue name is the identifier of the queue. If empty, the server will
// generate a unique name.
//
// The durable, autodelete, and exclusive flags control the durability and
// lifetime of the queue. If the durable flag is set, the queue will be
// persisted across a server restart. If the autodelete flag is set, the queue
// is deleted when the last consumer is cancelled. If the exclusive flag is set,
// the queue is deleted when the connection is closed.
//
// The noWait flag controls whether the server waits for confirmation from
// the queues that the queue has been created. If the flag is set, the server
// will not wait.
//
// The arguments map is a map of optional arguments that can be passed to the
// queue declaration. The keys and values are dependent on the underlying
// RabbitMQ server.
//
// The return value is the declared queue, or an error if the declaration fails.
func (rc *RabbitMQClient) CreateQueue(queueName string, durable, autodelete bool, exclusive bool, noWait bool, args map[string]interface{}) (amqp.Queue, error) {
	queue, err := rc.ch.QueueDeclare(queueName, durable, autodelete, exclusive, noWait, args)

	if err != nil {
		log.Error("Error creating a queue", err.Error())
		return amqp.Queue{}, err
	}

	return queue, nil
}

// CreateExchange declares a RabbitMQ exchange. If the exchange does not exist, it is
// created. If the exchange already exists, the exchange's properties are checked,
// but any mismatched properties do not cause an exception.
//
// The exchange name is the identifier of the exchange.
//
// The durable, autodelete, and internal flags control the durability and
// lifetime of the exchange. If the durable flag is set, the exchange will be
// persisted across a server restart. If the autodelete flag is set, the exchange
// is deleted when the last queue is unbound. If the internal flag is set, the
// exchange is only accessible from within the RabbitMQ server.
//
// The noWait flag controls whether the server waits for confirmation from
// the exchanges that the exchange has been created. If the flag is set, the server
// will not wait.
//
// The arguments map is a map of optional arguments that can be passed to the
// exchange declaration. The keys and values are dependent on the underlying
// RabbitMQ server.
//
// The return value is an error if the declaration fails.
func (rc *RabbitMQClient) CreateExchange(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args map[string]interface{}) error {
	if err := rc.ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args); err != nil {
		log.Error("Error binding a queue", err.Error())
		return err
	}

	return nil
}

// BindQueueWithExchange binds a queue with an exchange.
//
// The queue name is the identifier of the queue.
//
// The key is the routing key used to bind the queue to the exchange.
//
// The exchange name is the identifier of the exchange. If empty, the server will
// generate a unique name.
//
// The durable, autodelete, and exclusive flags control the durability and
// lifetime of the queue. If the durable flag is set, the queue will be
// persisted across a server restart. If the autodelete flag is set, the queue
// is deleted when the last consumer is cancelled. If the exclusive flag is set,
// the queue is deleted when the connection is closed.
//
// The noWait flag controls whether the server waits for confirmation from
// the queues that the queue has been created. If the flag is set, the server
// will not wait.
//
// The arguments map is a map of optional arguments that can be passed to the
// queue declaration. The keys and values are dependent on the underlying
// RabbitMQ server.
//
// The return value is an error if the declaration fails.
func (rc *RabbitMQClient) BindQueueWithExchange(queueName string, key string, exchangeName string, noWait bool, args map[string]interface{}) error {
	if err := rc.ch.QueueBind(queueName, key, exchangeName, noWait, args); err != nil {
		log.Error("Error binding a queue", err.Error())
		return err
	}

	return nil
}

// CreateMessgeChannel consumes a message from a RabbitMQ queue and returns a channel
// with the messages. If the connection to RabbitMQ is closed, the function will
// reconnect and try consuming again. If there is an error consuming the message,
// the function will return the error.
//
// The queue name is the identifier of the queue. If empty, the server will
// generate a unique name.
//
// The consumer name is the identifier of the consumer. If empty, the server will
// generate a unique name.
//
// The durable, autodelete, and exclusive flags control the durability and
// lifetime of the queue. If the durable flag is set, the queue will be
// persisted across a server restart. If the autodelete flag is set, the queue
// is deleted when the last consumer is cancelled. If the exclusive flag is set,
// the queue is deleted when the connection is closed.
//
// The noWait flag controls whether the server waits for confirmation from
// the queues that the queue has been created. If the flag is set, the server
// will not wait.
//
// The arguments map is a map of optional arguments that can be passed to the
// queue declaration. The keys and values are dependent on the underlying
// RabbitMQ server.
//
// The return value is an error if the declaration fails.
func (rc *RabbitMQClient) CreateMessgeChannel(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args map[string]interface{}) error {
	msgs, err := rc.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if err != nil {
		log.Error("Error conusming message", err.Error())
		return nil
	}
	rc.msgCh = msgs
	return nil
}

package main

import (
	"fmt"
	"os"
	"rabbit_security_micro/publisher/utils"
	"sync"

	"github.com/streadway/amqp"
)

func main() {
	url := os.Getenv("AMQP_URL")
	// desKey := []byte("12345678")
	var desKey []byte

	if url == "" {
		//url = "amqp://guest:guest@localhost:5672"
		url = "amqp://guest:guest@rabbitmq"

	}

	connection, err := amqp.Dial(url)

	if err != nil {
		panic("could not establish connection with RabbitMQ:" + err.Error())
	}

	channel, err := connection.Channel()
	keyChannel, err := connection.Channel()
	desChannel, err := connection.Channel()

	if err != nil {
		panic("could not open RabbitMQ channel:" + err.Error())
	}

	// We create an exahange that will bind to the queue to send and receive messages
	err = channel.ExchangeDeclare("events", "topic", true, false, false, false, nil)

	//Start of key exchange point
	err = keyChannel.ExchangeDeclare("KEY-EXCHANGE", "topic", true, false, false, false, nil)
	err = desChannel.ExchangeDeclare("des-key", "topic", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	_, err = desChannel.QueueDeclare("exchange-des", true, false, false, false, nil)
	// var network bytes.Buffer // Stand-in for a network connection
	// enc := gob.NewEncoder(&network)

	priv, pub := utils.GenerateKeyPair(2048)
	// fmt.Println("private key: ", priv)
	// enc.Encode(&pub)
	pubBytes := utils.PublicKeyToBytes(pub)

	// We create a message to be sent to the queue.
	// It has to be an instance of the aqmp publishing struct
	message := amqp.Publishing{
		Body: pubBytes,
	}

	_, err = keyChannel.QueueDeclare("exchange-queue", true, false, false, false, nil)
	_, err = keyChannel.QueueDeclare("exchange-des", true, false, false, false, nil)

	if err != nil {
		panic("error declaring the exchange queue: " + err.Error())
	}

	// We publish the message to the exahange we created earlier
	err = keyChannel.Publish("KEY-EXCHANGE", "exchange-queue", false, false, message)

	if err != nil {
		panic("error publishing a message to the queue:" + err.Error())
	}

	err = keyChannel.QueueBind("exchange-queue", "#", "KEY-EXCHANGE", false, nil)
	err = keyChannel.QueueBind("exchange-des", "#", "KEY-EXCHANGE", false, nil)

	err = desChannel.QueueBind("exchange-des", "#", "des-key", false, nil)
	if err != nil {
		panic("error binding the exchange queue: " + err.Error())
	}

	msgs1, err := keyChannel.Consume("exchange-des", "", false, false, false, false, nil)

	if err != nil {
		panic("error consuming the des queue: " + err.Error())
	}

	// We consume data from the queue named Test using the channel we created in go.
	// msgs1, err := keyChannel.Consume("exchange-des", "", false, false, false, false, nil)
	//
	// if err != nil {
	// 	panic("error consuming the exchange queue: " + err.Error())
	// }
	//
	fmt.Println("msgs1: ", msgs1)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for msg := range msgs1 {
			fmt.Println("DES KEY: " + string(msg.Body))
			decryptedKey := utils.DecryptWithPrivateKey(msg.Body, priv)
			fmt.Println("DECRYPTED DES KEY: " + string(decryptedKey))
			msg.Ack(false)
			wg.Done()
			if len(decryptedKey) > 0 {
				// wg.Done()
				desKey = decryptedKey
				return
			}
		}
	}()
	wg.Wait()

	// We create a queue named Test
	_, err = channel.QueueDeclare("test", true, false, false, false, nil)

	if err != nil {
		panic("error declaring the queue: " + err.Error())
	}

	// We bind the queue to the exchange to send and receive data from the queue
	err = channel.QueueBind("test", "#", "events", false, nil)

	if err != nil {
		panic("error binding to the queue: " + err.Error())
	}

	// We consume data from the queue named Test using the channel we created in go.
	msgs, err := channel.Consume("test", "", false, false, false, false, nil)

	if err != nil {
		panic("error consuming the queue: " + err.Error())
	}

	// We loop through the messages in the queue and print them in the console.
	// The msgs will be a go channel, not an amqp channel
	for msg := range msgs {
		decryptedText, _ := utils.DesDecrypt(msg.Body, desKey)
		fmt.Println("encrypted message received: " + string(msg.Body))
		fmt.Println("decrypted message: " + string(decryptedText))
		msg.Ack(false)
	}

	// We close the connection after the operation has completed.
	defer connection.Close()
}

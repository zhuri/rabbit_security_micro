package main

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"os"
	"rabbit_security_micro/publisher/utils"
	"strings"
	"sync"

	"github.com/streadway/amqp"
)

func main() {
	// Get the connection string from the environment var
	url := os.Getenv("AMQP_URL")
	desKey := []byte("12345678")
	var pub *rsa.PublicKey

	//If it doesnt exist, use the default connection string
	if url == "" {
		//url = "amqp://guest:guest@0.0.0.0:5672"
		url = "amqp://guest:guest@rabbitmq"
	}

	// Connect to the rabbitMQ instance
	connection, err := amqp.Dial(url)

	if err != nil {
		panic("could not establish connection with RabbitMQ:" + err.Error())
	}

	// Create a channel from the connection
	channel, err := connection.Channel()
	keyChannel, err := connection.Channel()
	desChannel, err := connection.Channel()

	if err != nil {
		panic("could not open RabbitMQ channel:" + err.Error())
	}

	// ExchangeDeclare - exchange point
	err = channel.ExchangeDeclare("events", "topic", true, false, false, false, nil)
	err = keyChannel.ExchangeDeclare("KEY-EXCHANGE", "topic", true, false, false, false, nil)
	err = desChannel.ExchangeDeclare("des-key", "topic", true, false, false, false, nil)

	if err != nil {
		panic(err)
	}

	_, err = keyChannel.QueueDeclare("exchange-queue", true, false, false, false, nil)
	_, err = desChannel.QueueDeclare("exchange-des", true, false, false, false, nil)
	// _, err = keyChannel.QueueDeclare("exchange-des", true, false, false, false, nil)

	if err != nil {
		panic("error declaring the exchange queue: " + err.Error())
	}

	err = keyChannel.QueueBind("exchange-queue", "#", "KEY-EXCHANGE", false, nil)
	if err != nil {
		panic("error binding the exchange queue: " + err.Error())
	}

	// We consume data from the queue named Test using the channel we created in go.
	msgs, err := keyChannel.Consume("exchange-queue", "", false, false, false, false, nil)

	if err != nil {
		panic("error consuming the exchange queue: " + err.Error())
	}

	// We loop through the messages in the queue and print them in the console.
	// The msgs will be a go channel, not an amqp channel
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for msg := range msgs {
			fmt.Println("key received: " + string(msg.Body))
			if msg.Body != nil || len(msg.Body) > 0 {
				pub = utils.BytesToPublicKey(msg.Body)
				fmt.Println("key received object : ", pub)
			}
			msg.Ack(false)
			wg.Done()
			// break
		}
	}()
	wg.Wait()
	fmt.Println("AFTER ROUTINE KEY: ", pub)
	encryptedDesKey := utils.EncryptWithPublicKey(desKey, pub)
	fmt.Println("Encrypted DES key: ", encryptedDesKey)

	// We create a queue message of amqp publishing instance

	message := amqp.Publishing{
		Body: encryptedDesKey,
	}
	// 	// //
	// 	// // // We publish the message to the EXCHANGE
	err = desChannel.Publish("des-key", "random-key", false, false, message)
	if err != nil {
		panic("error publishing a message to the exchange des queue:" + err.Error())
	}
	err = desChannel.QueueBind("exchange-des", "#", "des-key", false, nil)
	if err != nil {
		panic("error binding the exchange queue: " + err.Error())
	}

	// //Start of key exchange point
	// _, err = channel.QueueDeclare("KEY-EXCHANGE", true, false, false, false, nil)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// priv, pub := utils.GenerateKeyPair(128)
	// fmt.Println(priv, pub)

	// We create a queue message of amqp publishing instance
	// message := amqp.Publishing{
	// 	Body: []byte(string(pub)),
	// }

	// We publish the message to the EXCHANGE
	// err = channel.Publish("events", "random-key", false, false, message)

	//End of key exchange point

	// We create a queue named Test
	_, err = channel.QueueDeclare("test", true, false, false, false, nil)
	// _, err = keyChannel.QueueDeclare("exchange-queue", true, false, false, false, nil)

	if err != nil {
		panic("error declaring the queue: " + err.Error())
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		encryptedText, _ := utils.DesEncrypt([]byte(text), desKey)

		if strings.Compare("exit", text) == 0 {
			return
		}

		// We create a queue message of amqp publishing instance
		message := amqp.Publishing{
			Body: encryptedText,
		}

		// We publish the message to the EXCHANGE
		err = channel.Publish("events", "random-key", false, false, message)

		if err != nil {
			panic("error publishing a message to the queue:" + err.Error())
		}

		// We bind the queue to the exchange to send and receive data from the queue
		err = channel.QueueBind("test", "#", "events", false, nil)

		if err != nil {
			panic("error binding to the queue: " + err.Error())
		}
	}
}

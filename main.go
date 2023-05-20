package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type greetings struct {
	Message string `json:"message"`
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	crypto := filepath.Join(pwd, "crypto")

	awsKey := flag.String("aws-key", filepath.Join(crypto, "aws.key"), "AWS mTLS secret key")
	awsCert := flag.String("aws-cert", filepath.Join(crypto, "aws.crt"), "AWS mTLS certificate")
	awsCA := flag.String("aws-ca", filepath.Join(crypto, "aws-ca.pem"), "AWS mTLS CA")
	endpoint := flag.String("endpoint", "ssl://a1ya5rmdywas0n-ats.iot.eu-north-1.amazonaws.com:8883", "AWS MQTT endpoint to connect to")
	thingName := flag.String("thing-name", "GreenGuardianGateway1", "Thing name (for topic to publish too; invalid thing names are denied using the )")

	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*awsCert, *awsKey)
	if err != nil {
		panic(err)
	}

	ca, err := os.ReadFile(*awsCA)
	if err != nil {
		panic(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)

	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*endpoint)
	opts.SetClientID(*thingName)
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(1000)

	log.Println("Connected to", *endpoint)

	b, err := json.Marshal(greetings{
		Message: "Hello, world!",
	})
	if err != nil {
		panic(err)
	}

	publishTopic := path.Join("/gateways", *thingName, "messages")

	if token := client.Publish(publishTopic, 0, false, b); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Printf("Sent %s to %v", b, publishTopic)

	subscribeTopic := path.Join("/gateways", *thingName, "actions")

	log.Println("Subscribed to messages from", subscribeTopic)

	if token := client.Subscribe(
		subscribeTopic,
		0,
		func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("Received %s from %s", msg.Payload(), msg.Topic())
		},
	); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	select {}
}

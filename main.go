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
	"github.com/google/uuid"
)

type greeting struct {
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
	opts.SetClientID(uuid.New().String()[:22])
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(1000)

	log.Println("Connected to", *endpoint)

	b, err := json.Marshal(greeting{
		Message: "Hello, world!",
	})
	if err != nil {
		panic(err)
	}

	topic := path.Join("/gateways", *thingName, "greetings")

	if token := client.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		panic(err)
	}

	log.Printf("Sent %s to %v", b, topic)

	select {}
}

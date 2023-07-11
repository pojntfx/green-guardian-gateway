package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/green-guardian-gateway/pkg/services"
	uutils "github.com/pojntfx/green-guardian-gateway/pkg/utils"
	"github.com/pojntfx/r3map/pkg/utils"
)

func main() {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Define the destination of crypto files
	crypto := filepath.Join(pwd, "crypto")

	// Use command line flags to allow customization of certain parameters
	// For each setting, if not provided in command-line args, defaults are taken from environment variables.
	// Otherwise, a default value is used.
	laddr := flag.String("laddr", uutils.GetStringEnvOrDefault("LADDR", ":1337"), "Listen address")
	verbose := flag.Bool("verbose", uutils.GetBoolEnvOrDefault("VERBOSE", false), "Whether to enable verbose logging")

	// Define AWS key, cert and ca location or path
	awsKey := flag.String("aws-key", uutils.GetStringEnvOrDefault("AWS_KEY", filepath.Join(crypto, "key.pem")), "AWS mTLS secret key")
	awsCert := flag.String("aws-cert", uutils.GetStringEnvOrDefault("AWS_CERT", filepath.Join(crypto, "cert.pem")), "AWS mTLS certificate")
	awsCA := flag.String("aws-ca", uutils.GetStringEnvOrDefault("AWS_CA", filepath.Join(crypto, "ca.pem")), "AWS mTLS CA")

	// Define endpoint and thing name
	endpoint := flag.String("endpoint", uutils.GetStringEnvOrDefault("ENDPOINT", "ssl://ad218s2flbk57-ats.iot.eu-central-1.amazonaws.com:8883"), "AWS MQTT endpoint to connect to")
	thingName := flag.String("thing-name", uutils.GetStringEnvOrDefault("THING_NAME", "DEVICE-Device_1"), "Thing name (for topic to publish too; invalid thing names are denied using the )")

	// Parse all defined flags
	flag.Parse()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load AWS certificate and keys
	cert, err := tls.LoadX509KeyPair(*awsCert, *awsKey)
	if err != nil {
		panic(err)
	}

	// Load AWS CA
	ca, err := os.ReadFile(*awsCA)
	if err != nil {
		panic(err)
	}

	// Append AWS certificate to certificate pool
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)

	// Create a new TLS Config with root CAs from the pool and certificate
	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
	}

	// Define the client options in MQTT
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*endpoint)
	opts.SetClientID(*thingName)
	opts.SetTLSConfig(tlsConfig)

	// Create and connect the MQTT client
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(1000)

	log.Println("Connected to", *endpoint)

	// Create a new Gateway
	gateway := services.NewGateway(
		*verbose,
		ctx,
		client,
		*thingName,
	)

	errs := make(chan error)
	go func() {
		if err := services.WaitGateway(gateway); err != nil {
			errs <- err
		}
	}()

	if err := services.OpenGateway(gateway, ctx); err != nil {
		panic(err)
	}
	defer services.CloseGateway(gateway)

	clients := 0

	// Create a new registry which will manage remote procedure calls and manage peers.
	registry := rpc.NewRegistry(
		gateway,
		services.HubRemote{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
			// Callback when a client is connected or disconnected to update the number of active clients
			OnClientConnect: func(remoteID string) {
				clients++

				log.Printf("%v clients connected", clients)
			},
			OnClientDisconnect: func(remoteID string) {
				clients--

				log.Printf("%v clients connected", clients)
			},
		},
	)
	// Assign this Registry's peers to Gateway's peers
	gateway.Peers = registry.Peers

	// Start listening for TCP connections
	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	log.Println("Listening on", lis.Addr())

	// Accept new connections
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				if !utils.IsClosedErr(err) {
					log.Println("could not accept connection, continuing:", err)
				}

				continue
			}

			go func() {
				defer func() {
					// Close the connection
					_ = conn.Close()

					if err := recover(); err != nil {
						if !utils.IsClosedErr(err.(error)) {
							log.Printf("Client disconnected with error: %v", err)
						}
					}
				}()

				// Link the RPCs to the connection
				if err := registry.Link(conn); err != nil {
					panic(err)
				}
			}()
		}
	}()

	// Wait for any errors to occur
	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

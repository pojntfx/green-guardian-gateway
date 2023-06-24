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
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	crypto := filepath.Join(pwd, "crypto")

	laddr := flag.String("laddr", uutils.GetStringEnvOrDefault("LADDR", ":1337"), "Listen address")
	verbose := flag.Bool("verbose", uutils.GetBoolEnvOrDefault("VERBOSE", false), "Whether to enable verbose logging")
	awsKey := flag.String("aws-key", uutils.GetStringEnvOrDefault("AWS_KEY", filepath.Join(crypto, "key.pem")), "AWS mTLS secret key")
	awsCert := flag.String("aws-cert", uutils.GetStringEnvOrDefault("AWS_CERT", filepath.Join(crypto, "cert.pem")), "AWS mTLS certificate")
	awsCA := flag.String("aws-ca", uutils.GetStringEnvOrDefault("AWS_CA", filepath.Join(crypto, "ca.pem")), "AWS mTLS CA")
	endpoint := flag.String("endpoint", uutils.GetStringEnvOrDefault("ENDPOINT", "ssl://ad218s2flbk57-ats.iot.eu-central-1.amazonaws.com:8883"), "AWS MQTT endpoint to connect to")
	thingName := flag.String("thing-name", uutils.GetStringEnvOrDefault("THING_NAME", "DEVICE-Device_1"), "Thing name (for topic to publish too; invalid thing names are denied using the )")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	registry := rpc.NewRegistry(
		gateway,
		services.HubRemote{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
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
	gateway.Peers = registry.Peers

	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	log.Println("Listening on", lis.Addr())

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
					_ = conn.Close()

					if err := recover(); err != nil {
						if !utils.IsClosedErr(err.(error)) {
							log.Printf("Client disconnected with error: %v", err)
						}
					}
				}()

				if err := registry.Link(conn); err != nil {
					panic(err)
				}
			}()
		}
	}()

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

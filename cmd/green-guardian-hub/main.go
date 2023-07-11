package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/green-guardian-gateway/pkg/services"
	uutils "github.com/pojntfx/green-guardian-gateway/pkg/utils"
	"github.com/pojntfx/r3map/pkg/utils"
	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

var (
	errNoPeerFound = errors.New("no peer found")
)

func main() {
	// Define variables to get environment values or default ones if environment variables are not set
	baudDefault, err := uutils.GetIntEnvOrDefault("BAUD", 115200)
	if err != nil {
		panic(err)
	}
	baud := flag.Int("baud", baudDefault, "Baudrate to use to communicate with sensors and actuators")

	raddr := flag.String("raddr", uutils.GetStringEnvOrDefault("RADDR", "localhost:1337"), "Remote address")
	verbose := flag.Bool("verbose", uutils.GetBoolEnvOrDefault("VERBOSE", false), "Whether to enable verbose logging")

	defaultTempDefault, err := uutils.GetIntEnvOrDefault("DEFAULT_TEMPERATURE", 25)
	if err != nil {
		panic(err)
	}
	defaultTemperature := flag.Int("default-temperature", defaultTempDefault, "The default expected temperature")

	defaultMoistureDefault, err := uutils.GetIntEnvOrDefault("DEFAULT_MOISTURE", 30)
	if err != nil {
		panic(err)
	}
	defaultMoisture := flag.Int("default-moisture", defaultMoistureDefault, "The default expected moisture")

	measureIntervalDefault, err := uutils.GetDurationEnvOrDefault("MEASURE_INTERVAL", time.Second)
	if err != nil {
		panic(err)
	}
	measureInterval := flag.Duration("measure-interval", measureIntervalDefault, "Amount of time after which a new measurement is taken")

	measureTimeoutDefault, err := uutils.GetDurationEnvOrDefault("MEASURE_TIMEOUT", time.Second)
	if err != nil {
		panic(err)
	}
	measureTimeout := flag.Duration("measure-timeout", measureTimeoutDefault, "Amount of time after which it is assumed that a measurement has failed")

	// Define a JSON structure for each peripheral device
	fans := flag.String("fans", uutils.GetStringEnvOrDefault("FANS", `{"1": "/dev/ttyACM0"}`), "JSON description in the format { roomID: devicePath }")
	temperatureSensors := flag.String("temperature-sensors", uutils.GetStringEnvOrDefault("TEMPERATURE_SENSORS", `{"1": "/dev/ttyACM0"}`), "JSON description in the format { roomID: devicePath }")
	sprinklers := flag.String("sprinklers", uutils.GetStringEnvOrDefault("SPRINKLERS", `{"1": "/dev/ttyACM0"}`), "JSON description in the format { plantID: devicePath }")
	moistureSensors := flag.String("moisture-sensors", uutils.GetStringEnvOrDefault("MOISTURE_SENSORS", `{"1": "/dev/ttyACM0"}`), "JSON description in the format { roomID: devicePath }")

	// Mock for development and testing purposes
	mockDefault, err := uutils.GetIntEnvOrDefault("MOCK", 0)
	if err != nil {
		panic(err)
	}
	mock := flag.Int("mock", mockDefault, "If set to >1, mock temperature and moisture using buttons, sending the default value +- the value of this flag")

	// Parse command line flags
	flag.Parse()

	// Cancelable context for managing long-running go routines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Declare a map to hold fan devices, then unmarshal JSON into this map
	fanDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*fans), &fanDevices); err != nil {
		panic(err)
	}

	// Open connections for each fan device and store them in a map
	fanBindings := map[string]uutils.IoTee{}
	for roomID, dev := range fanDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		fanBindings[roomID] = uutils.NewIoTeeAdapter(it)
	}

	// Similarly for temperature sensors, sprinklers and moisture sensors
	temperatureSensorDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*temperatureSensors), &temperatureSensorDevices); err != nil {
		panic(err)
	}

	temperatureSensorBindings := map[string]uutils.IoTee{}
	for roomID, dev := range temperatureSensorDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		temperatureSensorBindings[roomID] = uutils.NewIoTeeAdapter(it)
	}

	sprinklerDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*sprinklers), &sprinklerDevices); err != nil {
		panic(err)
	}

	sprinklerBindings := map[string]uutils.IoTee{}
	for roomID, dev := range fanDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		sprinklerBindings[roomID] = uutils.NewIoTeeAdapter(it)
	}

	moistureSensorDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*moistureSensors), &moistureSensorDevices); err != nil {
		panic(err)
	}

	moistureSensorBindings := map[string]uutils.IoTee{}
	for roomID, dev := range moistureSensorDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		moistureSensorBindings[roomID] = uutils.NewIoTeeAdapter(it)
	}

	// Initialization of hub, the main service that communicates with sensors/actuators
	hub := services.NewHub(
		*verbose,
		ctx,

		fanBindings,
		temperatureSensorBindings,
		*defaultTemperature,

		sprinklerBindings,
		moistureSensorBindings,
		*defaultMoisture,

		*measureInterval,
		*measureTimeout,

		*mock,
	)

	// Registry where every connected device is registered
	ready := make(chan struct{})
	registry := rpc.NewRegistry(
		hub,
		services.GatewayRemote{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
			OnClientConnect: func(remoteID string) {
				ready <- struct{}{}
			},
		},
	)

	// Dial remote address
	conn, err := net.Dial("tcp", *raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Link RPCs and check for errors
	go func() {
		if err := registry.Link(conn); err != nil {
			if !utils.IsClosedErr(err) {
				panic(err)
			}
		}
	}()

	// Wait for connection
	<-ready

	log.Println("Connected to", conn.RemoteAddr())

	// Find the first peer
	var peer *services.GatewayRemote
	for _, candidate := range registry.Peers() {
		peer = &candidate

		break
	}

	if peer == nil {
		panic(errNoPeerFound)
	}

	// Error channel to handle errors in Go routines
	errs := make(chan error)
	go func() {
		if err := services.WaitHub(hub); err != nil {
			errs <- err
		}
	}()

	if err := services.OpenHub(hub, ctx, peer); err != nil {
		panic(err)
	}

	// Handling interrupt signal for shutdown
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)

		<-ch

		if *verbose {
			log.Println("Gracefully shutting down")

			go func() {
				ch := make(chan os.Signal, 1)
				signal.Notify(ch, os.Interrupt)

				<-ch

				log.Println("Forcefully exiting")

				os.Exit(1)
			}()
		}

		_ = services.CloseHub(hub, ctx, peer)

		os.Exit(1)
	}()

	// Error handling loop
	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

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
	"github.com/pojntfx/r3map/pkg/utils"
	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

var (
	errNoPeerFound = errors.New("no peer found")
)

func main() {
	baud := flag.Int("baud", 115200, "Baudrate to use to communicate with sensors and actuators")
	raddr := flag.String("raddr", "localhost:1337", "Remote address")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	defaultTemperature := flag.Int("default-temperature", 25, "The default expected temperature")
	defaultMoisture := flag.Int("default-moisture", 30, "The default expected moisture")
	measureInterval := flag.Duration("measure-interval", time.Second, "Amount of time after which a new measurement is taken")
	measureTimeout := flag.Duration("measure-timeout", time.Second, "Amount of time after which it is assumed that a measurement has failed")
	fans := flag.String("fans", `{"1": "/dev/ttyACM0"}`, "JSON description in the format { roomID: devicePath }")
	temperatureSensors := flag.String("temperature-sensors", `{"1": "/dev/ttyACM0"}`, "JSON description in the format { roomID: devicePath }")
	sprinklers := flag.String("sprinklers", `{"1": "/dev/ttyACM0"}`, "JSON description in the format { plantID: devicePath }")
	moistureSensors := flag.String("moisture-sensors", `{"1": "/dev/ttyACM0"}`, "JSON description in the format { roomID: devicePath }")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fanDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*fans), &fanDevices); err != nil {
		panic(err)
	}

	fanBindings := map[string]*iotee.IoTee{}
	for roomID, dev := range fanDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		fanBindings[roomID] = it
	}

	temperatureSensorDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*temperatureSensors), &temperatureSensorDevices); err != nil {
		panic(err)
	}

	temperatureSensorBindings := map[string]*iotee.IoTee{}
	for roomID, dev := range temperatureSensorDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		temperatureSensorBindings[roomID] = it
	}

	sprinklerDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*sprinklers), &sprinklerDevices); err != nil {
		panic(err)
	}

	sprinklerBindings := map[string]*iotee.IoTee{}
	for roomID, dev := range fanDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		sprinklerBindings[roomID] = it
	}

	moistureSensorDevices := map[string]string{}
	if err := json.Unmarshal([]byte(*moistureSensors), &moistureSensorDevices); err != nil {
		panic(err)
	}

	moistureSensorBindings := map[string]*iotee.IoTee{}
	for roomID, dev := range moistureSensorDevices {
		it := iotee.NewIoTee(dev, *baud)

		if err := it.Open(); err != nil {
			panic(err)
		}
		defer it.Close()

		moistureSensorBindings[roomID] = it
	}

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
	)

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

	conn, err := net.Dial("tcp", *raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		if err := registry.Link(conn); err != nil {
			if !utils.IsClosedErr(err) {
				panic(err)
			}
		}
	}()

	<-ready

	log.Println("Connected to", conn.RemoteAddr())

	var peer *services.GatewayRemote
	for _, candidate := range registry.Peers() {
		peer = &candidate

		break
	}

	if peer == nil {
		panic(errNoPeerFound)
	}

	errs := make(chan error)
	go func() {
		if err := services.WaitHub(hub); err != nil {
			errs <- err
		}
	}()

	if err := services.OpenHub(hub, ctx, peer); err != nil {
		panic(err)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)

		<-ch

		if *verbose {
			log.Println("Gracefully shutting down")
		}

		_ = services.CloseHub(hub, ctx, peer)

		os.Exit(1)
	}()

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

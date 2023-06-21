package services

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"sync"
	"time"

	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

var (
	ErrNoSuchRoom              = errors.New("no such room")
	ErrTemperatureReadTimedOut = errors.New("temperature read timed out")
)

type HubRemote struct {
	SetFanOn func(ctx context.Context, roomID string, on bool) error
	// SetSprinklerOn func(ctx context.Context, plantID string, on bool) error
}

type Hub struct {
	verbose bool

	ctx    context.Context
	cancel context.CancelFunc

	errs chan error

	fans               map[string]*iotee.IoTee
	temperatureSensors map[string]*iotee.IoTee

	defaultTemperature int

	measureInterval,
	measureTimeout time.Duration

	workerWg sync.WaitGroup
}

func NewHub(
	verbose bool,
	ctx context.Context,
	fans map[string]*iotee.IoTee,
	temperatureSensors map[string]*iotee.IoTee,
	defaultTemperature int,
	measureInterval,
	measureTimeout time.Duration,
) *Hub {
	cancellableCtx, cancel := context.WithCancel(ctx)

	return &Hub{
		verbose: verbose,

		ctx:    cancellableCtx,
		cancel: cancel,

		errs: make(chan error),

		fans:               fans,
		temperatureSensors: temperatureSensors,

		defaultTemperature: defaultTemperature,

		measureInterval: measureInterval,
		measureTimeout:  measureTimeout,
	}
}

func (w *Hub) SetFanOn(ctx context.Context, roomID string, on bool) error {
	if w.verbose {
		log.Printf("SetFanOn(roomID=%v, on=%v)", roomID, on)
	}

	fan, ok := w.fans[roomID]
	if !ok {
		return ErrNoSuchRoom
	}

	req := iotee.NewMessage(iotee.MessageTypeRGBLED, 4)

	intensity := byte(0)
	if on {
		intensity = 255
	}

	req.Data = []byte{intensity, 255, 0, 0}

	return fan.Transmit(&req)
}

func OpenHub(hub *Hub, ctx context.Context, gateway *GatewayRemote) error {
	roomIDs := []string{}
	for roomID := range hub.fans {
		roomIDs = append(roomIDs, roomID)
	}

	if err := gateway.RegisterFans(ctx, roomIDs); err != nil {
		return err
	}

	for roomID, temperatureSensor := range hub.temperatureSensors {
		hub.workerWg.Add(1)

		go func(roomID string, temperatureSensor *iotee.IoTee) {
			defer hub.workerWg.Done()

			for {
				select {
				case <-hub.ctx.Done():
					return
				default:
					req := iotee.NewMessage(iotee.MessageTypeTempReq, 0)

					if err := temperatureSensor.Transmit(&req); err != nil {
						hub.errs <- err

						return
					}

					res := temperatureSensor.ReceiveWithTimeout(hub.measureTimeout)
					if res == nil {
						hub.errs <- ErrTemperatureReadTimedOut

						return
					}

					if err := gateway.ForwardTemperatureMeasurement(ctx, roomID, int(float32(binary.BigEndian.Uint32(res.Data[0:4]))/100.0), hub.defaultTemperature); err != nil {
						hub.errs <- err

						return
					}

					time.Sleep(hub.measureInterval)
				}
			}
		}(roomID, temperatureSensor)
	}

	return nil
}

func WaitHub(hub *Hub) error {
	for err := range hub.errs {
		if err != nil {
			return err
		}
	}

	hub.workerWg.Wait()

	return nil
}

func CloseHub(hub *Hub, ctx context.Context, gateway *GatewayRemote) error {
	roomIDs := []string{}
	for roomID := range hub.fans {
		roomIDs = append(roomIDs, roomID)
	}

	if err := gateway.UnregisterFans(ctx, roomIDs); err != nil {
		return err
	}

	hub.cancel()

	close(hub.errs)

	hub.workerWg.Wait()

	return nil
}

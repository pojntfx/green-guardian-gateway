package services

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/pojntfx/green-guardian-gateway/pkg/utils"
	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

var (
	ErrNoSuchRoom  = errors.New("no such room")
	ErrNoSuchPlant = errors.New("no such plant")

	ErrTemperatureReadTimedOut = errors.New("temperature read timed out")
	ErrMoistureReadTimedOut    = errors.New("moisture read timed out")
)

type HubRemote struct {
	SetFanOn       func(ctx context.Context, roomID string, on bool) error
	SetSprinklerOn func(ctx context.Context, plantID string, on bool) error
}

type Hub struct {
	verbose bool

	ctx    context.Context
	cancel context.CancelFunc

	errs chan error

	fans               map[string]utils.IoTee
	temperatureSensors map[string]utils.IoTee

	defaultTemperature int

	sprinklers      map[string]utils.IoTee
	moistureSensors map[string]utils.IoTee

	defaultMoisture int

	measureInterval,
	measureTimeout time.Duration

	measureLock sync.Mutex

	workerWg sync.WaitGroup

	mock int
}

func NewHub(
	verbose bool,
	ctx context.Context,

	fans map[string]utils.IoTee,
	temperatureSensors map[string]utils.IoTee,
	defaultTemperature int,

	sprinklers map[string]utils.IoTee,
	moistureSensors map[string]utils.IoTee,
	defaultMoisture int,

	measureInterval,
	measureTimeout time.Duration,

	mock int,
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

		sprinklers:      sprinklers,
		moistureSensors: moistureSensors,

		defaultMoisture: defaultMoisture,

		measureInterval: measureInterval,
		measureTimeout:  measureTimeout,

		mock: mock,
	}
}

// SetFanOn turns the specified fan on or off.
func (w *Hub) SetFanOn(ctx context.Context, roomID string, on bool) error {
	if w.verbose {
		// Log the function call if verbose logging is enabled.
		log.Printf("SetFanOn(roomID=%v, on=%v)", roomID, on)
	}

	// Find the fan in the map using the roomID.
	fan, ok := w.fans[roomID]
	if !ok {
		// If the fan doesn't exist, return an error.
		return ErrNoSuchRoom
	}

	// Create a new IoT message.
	req := iotee.NewMessage(iotee.MessageTypeRGBLED, 4)

	// Set the intensity depending on the 'on' variable.
	intensity := byte(0)
	if on {
		intensity = 255
	}

	// Set the data of the message.
	req.Data = []byte{intensity, 255, 0, 0}

	// Transmit the message using the fan.
	return fan.Transmit(&req)
}

// SetSprinklerOn turns the specified sprinkler on or off.
func (w *Hub) SetSprinklerOn(ctx context.Context, roomID string, on bool) error {
	if w.verbose {
		// Log the function call if verbose logging is enabled.
		log.Printf("SetSprinklerOn(roomID=%v, on=%v)", roomID, on)
	}

	// Find the sprinkler in the map using the roomID.
	sprinkler, ok := w.sprinklers[roomID]
	if !ok {
		// If the sprinkler doesn't exist, return an error.
		return ErrNoSuchRoom
	}

	// Create a new IoT message.
	req := iotee.NewMessage(iotee.MessageTypeRGBLED, 4)

	// Set the intensity depending on the 'on' variable.
	intensity := byte(0)
	if on {
		intensity = 255
	}

	// Set the data of the message.
	req.Data = []byte{intensity, 0, 255, 0}

	// Transmit the message using the sprinkler.
	return sprinkler.Transmit(&req)
}

// OpenHub fires up the hub by registering fans and sprinklers with the gateway, and setting up temperature and moisture sensor handling.
func OpenHub(hub *Hub, ctx context.Context, gateway *GatewayRemote) error {
	// Prepare the list of room IDs
	roomIDs := []string{}
	for roomID := range hub.fans {
		roomIDs = append(roomIDs, roomID)
	}

	// Register fans if any.
	if len(hub.fans) > 0 {
		if err := gateway.RegisterFans(ctx, roomIDs); err != nil {
			return err
		}
	}

	// Register sprinklers if any.
	if len(hub.sprinklers) > 0 {
		if err := gateway.RegisterSprinklers(ctx, roomIDs); err != nil {
			return err
		}
	}

	// If mock mode is on, setup data handlers accordingly.
	if hub.mock > 0 {
		// When mocking, we treat all temperatures as the same
		for roomID, temperatureSensor := range hub.temperatureSensors {
			hub.workerWg.Add(1)

			go temperatureSensor.RxPump()

			go func(roomID string, temperatureSensor utils.IoTee) {
				defer hub.workerWg.Done()

				for {
					select {
					case <-hub.ctx.Done():
						return

					case msg := <-temperatureSensor.RxChan():
						if msg.MsgType == iotee.MessageTypeButton {
							switch msg.Data[0] {
							// Top left
							case 'B':
								if err := gateway.ForwardTemperatureMeasurement(ctx, roomID, hub.defaultTemperature+hub.mock, hub.defaultTemperature); err != nil {
									hub.errs <- err

									return
								}

							// Bottom left
							case 'Y':
								if err := gateway.ForwardTemperatureMeasurement(ctx, roomID, hub.defaultTemperature-hub.mock, hub.defaultTemperature); err != nil {
									hub.errs <- err

									return
								}

							// Top right
							case 'A':
								if err := gateway.ForwardMoistureMeasurement(ctx, roomID, hub.defaultMoisture-hub.mock, hub.defaultMoisture); err != nil {
									hub.errs <- err

									return
								}

							// Bottom right
							case 'X':
								if err := gateway.ForwardMoistureMeasurement(ctx, roomID, hub.defaultMoisture+hub.mock, hub.defaultMoisture); err != nil {
									hub.errs <- err

									return
								}
							}
						}
					}
				}
			}(roomID, temperatureSensor)
		}

		return nil
	}

	// Loop over all temperature sensors present in hub
	for roomID, temperatureSensor := range hub.temperatureSensors {
		hub.workerWg.Add(1) // increment the WaitGroup counter by one

		// Spin off a goroutine for each temperature sensor
		go func(roomID string, temperatureSensor utils.IoTee) {
			defer hub.workerWg.Done() // called at the end to notify that this goroutine is done

			for {
				select {
				// end the goroutine if the context signals done
				case <-hub.ctx.Done():
					return
				default:
					hub.measureLock.Lock() // use a lock to ensure safe concurrent access

					// create a new temperature request message
					req := iotee.NewMessage(iotee.MessageTypeTempReq, 0)

					// send a temperature request to the sensor
					if err := temperatureSensor.Transmit(&req); err != nil {
						hub.errs <- err // if there's an error, send it to the errors channel

						hub.measureLock.Unlock() // release the lock

						return
					}

					// receive the response from the temperature sensor with a timeout
					res := temperatureSensor.ReceiveWithTimeout(hub.measureTimeout)
					if res == nil {
						hub.errs <- ErrTemperatureReadTimedOut // if there's a timeout, send an error to the errors channel

						hub.measureLock.Unlock() // release the lock

						return
					}

					hub.measureLock.Unlock() // release the lock

					// forward the result to the gateway
					if err := gateway.ForwardTemperatureMeasurement(ctx, roomID, int(float32(binary.BigEndian.Uint32(res.Data[0:4]))/100.0), hub.defaultTemperature); err != nil {
						hub.errs <- err // if there's an error, send it to the errors channel

						return
					}

					time.Sleep(hub.measureInterval) // sleep for the duration of the measurement interval
				}
			}
		}(roomID, temperatureSensor)
	}

	// Loop over all moisture sensors present in hub
	for plantID, moistureSensor := range hub.moistureSensors {
		hub.workerWg.Add(1) // increment the WaitGroup counter by one

		// Spin off a goroutine for each moisture sensor
		go func(plantID string, moistureSensor utils.IoTee) {
			defer hub.workerWg.Done() // called at the end to notify that this goroutine is done

			for {
				select {
				// end the goroutine if the context signals done
				case <-hub.ctx.Done():
					return
				default:
					hub.measureLock.Lock() // use a lock to ensure safe concurrent access

					// create a new moisture request message
					req := iotee.NewMessage(iotee.MessageTypeHumReq, 0)

					// send a humidity request to the sensor
					if err := moistureSensor.Transmit(&req); err != nil {
						hub.errs <- err // if there's an error, send it to the errors channel

						hub.measureLock.Unlock() // release the lock

						return
					}

					// receive the response from the moisture sensor with a timeout
					res := moistureSensor.ReceiveWithTimeout(hub.measureTimeout)
					if res == nil {
						hub.errs <- ErrMoistureReadTimedOut // if there's a timeout, send an error to the errors channel

						hub.measureLock.Unlock() // release the lock

						return
					}

					hub.measureLock.Unlock() // release the lock

					// forward the result to the gateway
					if err := gateway.ForwardMoistureMeasurement(ctx, plantID, int(float32(binary.BigEndian.Uint32(res.Data[0:4]))/100.0), hub.defaultMoisture); err != nil {
						hub.errs <- err // if there's an error, send it to the errors channel

						return
					}

					time.Sleep(hub.measureInterval) // sleep for the duration of the measurement interval
				}
			}
		}(plantID, moistureSensor)
	}

	return nil
}

// WaitHub waits for the completion of the goroutines running in the hub.
func WaitHub(hub *Hub) error {
	// Check for any error from the error channel.
	for err := range hub.errs {
		if err != nil {
			return err
		}
	}

	// Wait for all goroutines to complete.
	hub.workerWg.Wait()

	return nil
}

// CloseHub performs cleanup operations on the hub such as unregistering fans and sprinklers and closing channels.
func CloseHub(hub *Hub, ctx context.Context, gateway *GatewayRemote) error {
	// Prepare the list of room IDs.
	roomIDs := []string{}
	for roomID := range hub.fans {
		roomIDs = append(roomIDs, roomID)
	}

	// Unregister fans from the gateway.
	if err := gateway.UnregisterFans(ctx, roomIDs); err != nil {
		return err
	}

	// Unregister sprinklers from the gateway.
	if err := gateway.UnregisterSprinklers(ctx, roomIDs); err != nil {
		return err
	}

	// Cancel the hub context.
	hub.cancel()

	// Close the error channel.
	close(hub.errs)

	// Wait for all worker goroutines to complete.
	hub.workerWg.Wait()

	return nil
}

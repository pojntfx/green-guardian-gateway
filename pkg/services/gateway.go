package services

import (
	"context"
	"encoding/json"
	"log"
	"path"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	mqttapi "github.com/pojntfx/green-guardian-gateway/pkg/api/mqtt"
)

type GatewayRemote struct {
	RegisterFans                  func(ctx context.Context, roomIDs []string) error
	UnregisterFans                func(ctx context.Context, roomIDs []string) error
	ForwardTemperatureMeasurement func(ctx context.Context, roomID string, measurement, defaultValue int) error

	RegisterSprinklers         func(ctx context.Context, plantIDs []string) error
	UnregisterSprinklers       func(ctx context.Context, plantIDs []string) error
	ForwardMoistureMeasurement func(ctx context.Context, plantID string, measurement, defaultValue int) error
}

type Gateway struct {
	verbose bool

	errs chan error

	broker    mqtt.Client
	thingName string

	fans     map[string]string
	fansLock sync.Mutex

	sprinklers     map[string]string
	sprinklersLock sync.Mutex

	Peers func() map[string]HubRemote
}

func NewGateway(
	verbose bool,
	ctx context.Context,
	broker mqtt.Client,
	thingName string,
) *Gateway {
	return &Gateway{
		verbose: verbose,

		errs: make(chan error),

		fans: map[string]string{},

		sprinklers: map[string]string{},

		broker:    broker,
		thingName: thingName,
	}
}

// RegisterFans method registers the rooms to the fans
func (w *Gateway) RegisterFans(ctx context.Context, roomIDs []string) error {
	if w.verbose {
		log.Printf("RegisterFans(roomIDs=%v)", roomIDs)
	}

	// Get the ID of the peer from the context
	peerID := rpc.GetRemoteID(ctx)

	// Lock the fans data preventing race condition
	w.fansLock.Lock()
	defer w.fansLock.Unlock()

	for _, roomID := range roomIDs {
		// Register the fan to the room
		w.fans[roomID] = peerID
	}

	return nil
}

// UnregisterFans method unregisters the rooms from the fans
func (w *Gateway) UnregisterFans(ctx context.Context, roomIDs []string) error {
	if w.verbose {
		log.Printf("UnregisterFans(roomIDs=%v)", roomIDs)
	}

	// Lock the fans data preventing race condition
	w.fansLock.Lock()
	defer w.fansLock.Unlock()

	for _, roomID := range roomIDs {
		// Unregister the fan from the room
		delete(w.fans, roomID)
	}

	return nil
}

// RegisterSprinklers method registers the plants to the sprinklers
func (w *Gateway) RegisterSprinklers(ctx context.Context, plantIDs []string) error {
	if w.verbose {
		log.Printf("RegisterSprinklers(plantIDs=%v)", plantIDs)
	}

	// Get the ID of the peer from the context
	peerID := rpc.GetRemoteID(ctx)

	// Lock the sprinklers data preventing race condition
	w.sprinklersLock.Lock()
	defer w.sprinklersLock.Unlock()

	for _, plantID := range plantIDs {
		// Register the sprinkler to the plant
		w.sprinklers[plantID] = peerID
	}

	return nil
}

// UnregisterSprinklers unregisters the plants from the sprinklers
func (w *Gateway) UnregisterSprinklers(ctx context.Context, plantIDs []string) error {
	if w.verbose {
		log.Printf("UnregisterSpriklers(plantIDs=%v)", plantIDs)
	}

	// Lock the sprinklers data preventing race condition
	w.sprinklersLock.Lock()
	defer w.sprinklersLock.Unlock()

	for _, plantID := range plantIDs {
		// Unregister the sprinkler from the plant
		delete(w.sprinklers, plantID)
	}

	return nil
}

// ForwardTemperatureMeasurement function is used to forward temperature measurements to the broker.
func (w *Gateway) ForwardTemperatureMeasurement(ctx context.Context, roomID string, measurement, defaultValue int) error {
	if w.verbose {
		log.Printf("ForwardTemperatureMeasurement(roomIDs=%v, measurement=%v, defaultValue=%v)", roomID, measurement, defaultValue)
	}

	// Marshal the measurement into a JSON format
	msg, err := json.Marshal(mqttapi.TemperatureMeasurement{
		Measurement:  measurement,
		DefaultValue: defaultValue,
	})
	if err != nil {
		return err
	}

	// Publish the measurement to the broker
	if token := w.broker.Publish(
		path.Join("/gateways", w.thingName, "rooms", roomID, "temperature"),
		0,
		false,
		msg,
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// ForwardMoistureMeasurement function is used to forward moisture measurements to the broker.
func (w *Gateway) ForwardMoistureMeasurement(ctx context.Context, plantID string, measurement, defaultValue int) error {
	if w.verbose {
		log.Printf("ForwardMoistureMeasurement(plantIDs=%v, measurement=%v, defaultValue=%v)", plantID, measurement, defaultValue)
	}

	// Marshal the measurement into a JSON format
	msg, err := json.Marshal(mqttapi.MoistureMeasurement{
		Measurement:  measurement,
		DefaultValue: defaultValue,
	})
	if err != nil {
		return err
	}

	// Publish the measurement to the broker
	if token := w.broker.Publish(
		path.Join("/gateways", w.thingName, "plants", plantID, "moisture"),
		0,
		false,
		msg,
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// OpenGateway function initializes gateway functionality by subscribing to fan and sprinkler MQTT topics.
func OpenGateway(gateway *Gateway, ctx context.Context) error {
	// Subscribe to fan topic
	if token := gateway.broker.Subscribe(
		path.Join("/gateways", gateway.thingName, "rooms", "+", "fan"),
		0,
		// Function to be called when a message on the fan topic is received
		func(client mqtt.Client, msg mqtt.Message) {
			gateway.fansLock.Lock()         // Lock to prevent concurrent modification
			defer gateway.fansLock.Unlock() // Unlock once finished

			basePath, _ := path.Split(msg.Topic())

			roomID := path.Base(basePath)

			// Check if fan exists for room
			peerID, ok := gateway.fans[roomID]
			if !ok {
				gateway.errs <- ErrNoSuchRoom

				return
			}

			// Get Hub for fan
			hub, ok := gateway.Peers()[peerID]
			if !ok {
				gateway.errs <- ErrNoSuchRoom

				return
			}

			// Parse FanState from message
			fanState := &mqttapi.FanState{}
			if err := json.Unmarshal(msg.Payload(), &fanState); err != nil {
				gateway.errs <- err

				return
			}

			// Attempt to turn fan on or off
			if err := hub.SetFanOn(ctx, roomID, fanState.On); err != nil {
				gateway.errs <- err

				return
			}
		},
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// Similar to above, subscribe to sprinkler topic
	if token := gateway.broker.Subscribe(
		path.Join("/gateways", gateway.thingName, "plants", "+", "sprinkler"),
		0,
		func(client mqtt.Client, msg mqtt.Message) {
			gateway.sprinklersLock.Lock()
			defer gateway.sprinklersLock.Unlock()

			basePath, _ := path.Split(msg.Topic())

			plantID := path.Base(basePath)

			// Check if sprinkler exists for plant
			peerID, ok := gateway.sprinklers[plantID]
			if !ok {
				gateway.errs <- ErrNoSuchPlant

				return
			}

			// Get Hub for sprinkler
			hub, ok := gateway.Peers()[peerID]
			if !ok {
				gateway.errs <- ErrNoSuchPlant

				return
			}

			// Parse SprinklerState from message
			sprinklerState := &mqttapi.SprinklerState{}
			if err := json.Unmarshal(msg.Payload(), &sprinklerState); err != nil {
				gateway.errs <- err

				return
			}

			// Attempt to turn sprinkler on or off
			if err := hub.SetSprinklerOn(ctx, plantID, sprinklerState.On); err != nil {
				gateway.errs <- err

				return
			}
		},
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// If everything went fine, return nil
	return nil
}

// WaitGateway is a helper function to handle errors from the gateway.
func WaitGateway(gateway *Gateway) error {
	for err := range gateway.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// CloseGateway function stops the gateway operation by unsubscribing from the MQTT topics and closing the error channel.
func CloseGateway(gateway *Gateway) error {
	// Unsubscribe from fan topic
	if token := gateway.broker.Unsubscribe(
		path.Join("/gateways", gateway.thingName, "rooms", "+", "fan"),
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// Unsubscribe from sprinkler topic
	if token := gateway.broker.Unsubscribe(
		path.Join("/gateways", gateway.thingName, "rooms", "+", "sprinkler"),
	); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// Close error channel
	close(gateway.errs)

	return nil
}

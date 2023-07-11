package services

//go:generate mockgen --destination gateway_mocks.go --package services --build_flags=--mod=mod github.com/eclipse/paho.mqtt.golang Client,Token

import (
	"context"
	"path"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/pojntfx/dudirekta/pkg/rpc"
)

// TestRegisterFans is a testing function that checks if the fans associated
// with given IDs are registered properly.
// Registration errors and issues with assigned IDs are reported as test failures.
func TestRegisterFans(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	gateway := NewGateway(false, ctx, nil, "TestThing")
	roomIDs := []string{"Room1", "Room2", "Room3"}

	if err := gateway.RegisterFans(ctx, roomIDs); err != nil {
		t.Fatalf("unexpected error during RegisterFans: %v", err)
	}

	for _, id := range roomIDs {
		if _, ok := gateway.fans[id]; !ok {
			t.Fatalf("fan with id %s was not registered", id)
		}
	}
}

// TestUnregisterFans is a function that tests the unregistration of fans,
// ensuring that the fans with the given IDs can be successfully unregistered.
// In case a fan ID is not unregistered properly, a test failure occurs.
func TestUnregisterFans(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	gateway := NewGateway(false, ctx, nil, "TestThing")
	roomIDs := []string{"Room1", "Room2", "Room3"}

	if err := gateway.RegisterFans(ctx, roomIDs); err != nil {
		t.Fatalf("unexpected error during RegisterFans: %v", err)
	}

	if err := gateway.UnregisterFans(ctx, roomIDs); err != nil {
		t.Fatalf("unexpected error during UnregisterFans: %v", err)
	}

	for _, id := range roomIDs {
		if _, ok := gateway.fans[id]; ok {
			t.Fatalf("fan with id %s was not unregistered", id)
		}
	}
}

// TestRegisterSprinklers is a function designed to test the registration
// of sprinklers by ID.
// Potential registration errors and issues with the assigned IDs
// result in a test failure.
func TestRegisterSprinklers(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")
	gateway := NewGateway(false, ctx, nil, "TestThing")
	plantIDs := []string{"Plant1", "Plant2", "Plant3"}

	if err := gateway.RegisterSprinklers(ctx, plantIDs); err != nil {
		t.Fatalf("unexpected error during RegisterSprinklers: %v", err)
	}

	for _, id := range plantIDs {
		if _, ok := gateway.sprinklers[id]; !ok {
			t.Fatalf("sprinkler with id %s was not registered", id)
		}
	}
}

// TestUnregisterSprinklers function tests server functionality for unregistering
// sprinklers based on given IDs.
// Fails the test if any sprinkler ID wasn't unregistered properly.
func TestUnregisterSprinklers(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")
	gateway := NewGateway(false, ctx, nil, "TestThing")
	plantIDs := []string{"Plant1", "Plant2", "Plant3"}

	if err := gateway.RegisterSprinklers(ctx, plantIDs); err != nil {
		t.Fatalf("unexpected error during RegisterSprinklers: %v", err)
	}

	if err := gateway.UnregisterSprinklers(ctx, plantIDs); err != nil {
		t.Fatalf("unexpected error during UnregisterSprinklers: %v", err)
	}

	for _, id := range plantIDs {
		if _, ok := gateway.sprinklers[id]; ok {
			t.Fatalf("sprinkler with id %s was not unregistered", id)
		}
	}
}

// TestForwardTemperatureMeasurement is a function that tests the ability
// of the gateway to forward temperature measurements.
// If there is an issue with forwarding the temperature measurement
// to the respective room, the test case fails.
func TestForwardTemperatureMeasurement(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroker := NewMockClient(ctrl)
	mockToken := NewMockToken(ctrl)

	mockToken.EXPECT().Wait().Return(true)
	mockToken.EXPECT().Error().Return(nil)

	gateway := NewGateway(false, ctx, mockBroker, "TestThing")
	roomID := "Room1"
	measurement := 25
	defaultValue := 20

	mockBroker.EXPECT().Publish(
		path.Join("/gateways", gateway.thingName, "rooms", roomID, "temperature"),
		byte(0),
		false,
		gomock.Any(),
	).Return(mockToken).Times(1)

	if err := gateway.ForwardTemperatureMeasurement(ctx, roomID, measurement, defaultValue); err != nil {
		t.Fatalf("unexpected error during ForwardTemperatureMeasurement: %v", err)
	}
}

// TestForwardMoistureMeasurement tests if the gateway is capable of
// forwarding the moisture measurement to the respective plant.
// If there is any error during the forwarding of moisture measurement,
// a test failure is reported.
func TestForwardMoistureMeasurement(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroker := NewMockClient(ctrl)
	mockToken := NewMockToken(ctrl)

	mockToken.EXPECT().Wait().Return(true)
	mockToken.EXPECT().Error().Return(nil)

	gateway := NewGateway(false, ctx, mockBroker, "TestThing")
	plantID := "Plant1"
	measurement := 35
	defaultValue := 30

	mockBroker.EXPECT().Publish(
		path.Join("/gateways", gateway.thingName, "plants", plantID, "moisture"),
		byte(0),
		false,
		gomock.Any(),
	).Return(mockToken).Times(1)

	if err := gateway.ForwardMoistureMeasurement(ctx, plantID, measurement, defaultValue); err != nil {
		t.Fatalf("unexpected error during ForwardMoistureMeasurement: %v", err)
	}
}

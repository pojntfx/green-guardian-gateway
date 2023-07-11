package services

//go:generate mockgen --destination gateway_mocks.go --package services --build_flags=--mod=mod github.com/eclipse/paho.mqtt.golang Client,Token

import (
	"context"
	"path"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/pojntfx/dudirekta/pkg/rpc"
)

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

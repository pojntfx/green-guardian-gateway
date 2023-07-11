package services

import (
	"context"
	"testing"

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

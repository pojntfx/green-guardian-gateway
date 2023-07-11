package services

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/green-guardian-gateway/pkg/utils"
	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

//go:generate mockgen --destination hub_mocks.go --package services --build_flags=--mod=mod github.com/pojntfx/green-guardian-gateway/pkg/utils IoTee

func TestSetFanOn(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFan := NewMockIoTee(ctrl)

	hub := NewHub(false, ctx, map[string]utils.IoTee{"Room1": mockFan}, nil, 0, nil, nil, 0, 0, 0, 0)

	roomID := "Room1"
	on := true
	expectedData := []byte{255, 255, 0, 0}

	mockFan.EXPECT().Transmit(&iotee.Message{
		MsgType: iotee.MessageTypeRGBLED,
		DataLen: 4,
		Data:    expectedData,
	}).Return(nil).Times(1)

	if err := hub.SetFanOn(ctx, roomID, on); err != nil {
		t.Fatalf("unexpected error during SetFanOn: %v", err)
	}
}

func TestSetSprinklerOn(t *testing.T) {
	ctx := context.WithValue(context.Background(), rpc.RemoteIDContextKey, "testremote")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSprinkler := NewMockIoTee(ctrl)

	hub := NewHub(false, ctx, nil, nil, 0, map[string]utils.IoTee{"Plant1": mockSprinkler}, nil, 0, 0, 0, 0)

	roomID := "Plant1"
	on := true
	expectedData := []byte{255, 0, 255, 0}

	mockSprinkler.EXPECT().Transmit(&iotee.Message{
		MsgType: iotee.MessageTypeRGBLED,
		DataLen: 4,
		Data:    expectedData,
	}).Return(nil).Times(1)

	if err := hub.SetSprinklerOn(ctx, roomID, on); err != nil {
		t.Fatalf("unexpected error during SetSprinklerOn: %v", err)
	}
}

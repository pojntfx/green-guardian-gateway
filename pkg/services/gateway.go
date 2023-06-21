package services

import "context"

type GatewayRemote struct {
	RegisterFans                  func(ctx context.Context, roomIDs []string) error
	ForwardTemperatureMeasurement func(ctx context.Context, roomID string, measurement, defaultValue int) error

	// RegisterSprinklers         func(ctx context.Context, plantIDs []string) error
	// ForwardMoistureMeasurement func(ctx context.Context, plantID string, measurement, defaultValue int) error
}

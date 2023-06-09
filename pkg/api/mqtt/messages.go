package mqtt

type FanState struct {
	On bool `json:"on"`
}

type SprinklerState = FanState

type TemperatureMeasurement struct {
	Measurement  int `json:"measurement"`
	DefaultValue int `json:"default"`
}

type MoistureMeasurement = TemperatureMeasurement

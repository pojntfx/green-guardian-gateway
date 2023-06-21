package mqtt

type FanState struct {
	On bool `json:"on"`
}

type TemperatureMeasurement struct {
	Measurement  int `json:"measurement"`
	DefaultValue int `json:"defaultValue"`
}

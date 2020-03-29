package utils

import (
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/protobufs/go/openiot/sensor"
	"github.com/stretchr/testify/assert"
)

func TestExtractAllValues(t *testing.T) {
	req := &openiot.KeyExchangeRequest{
		DhG: 111,
		DhA: []uint32{1, 2, 3},
	}
	expected := map[string]interface{}{
		"dh_g":            uint64(111),
		"dh_a":            []uint32{1, 2, 3},
		"dh_p":            uint64(0),
		"encryption_type": openiot.EncryptionType_PLAIN,
	}

	results := ExtractAllNameValuesFromProtobuf(req)
	assert.Equal(t, expected, results)
}

func TestExtractAllValuesEmbedded(t *testing.T) {
	req := &sensor.MultiSensorStatus{
		Temperature: &sensor.Temperature{
			ValueC: 10.1,
			ValueF: 99.9,
		},
		Humidity: &sensor.Humidity{
			RelativePercent: 50,
		},
	}
	expected := map[string]interface{}{
		"temperature.value_c":       float32(10.1),
		"temperature.value_f":       float32(99.9),
		"humidity.relative_percent": uint32(50),
	}

	results := ExtractAllNameValuesFromProtobuf(req)
	assert.Equal(t, expected, results)
}

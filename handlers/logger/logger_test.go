package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/protobufs/go/openiot/sensor"
)

func TestExtractNamesValuesFlatStruct(t *testing.T) {
	req := &openiot.KeyExchangeRequest{
		DhP: 99999,
		DhA: []uint32{1, 2, 3, 4, 5},
	}

	expected := []string{
		"dh_p: 99999",
		"dh_g: 0",
		"dh_a: [1 2 3 4 5]",
		"encryption_type: PLAIN",
	}
	results := extractNameValuesFromMessage(req)

	assert.Equal(t, expected, results)
}

func TestExtractNamesValuesEmbeddedStruct(t *testing.T) {
	req := &sensor.MultiSensorStatus{
		Temperature: &sensor.Temperature{
			ValueC: 10.1,
			ValueF: 99.9,
		},
		Humidity: &sensor.Humidity{
			RelativePercent: 50,
		},
	}

	expected := []string{
		"temperature: value_c:10.1 value_f:99.9 ",
		"humidity: relative_percent:50 ",
		// Battery is not defined, so should not be included
	}
	results := extractNameValuesFromMessage(req)

	assert.Equal(t, expected, results)
}

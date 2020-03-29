package influx

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gopkg.in/yaml.v2"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/utils"
)

const handlerName = "influxdb"

type kv map[string]interface{}

type influxDbConfig struct {
	Addr               string
	Username           string
	Password           string
	UserAgent          string
	InsecureSkipVerify bool

	Database string
	Enabled  bool
}

type deviceHandler struct {
	config influxDbConfig
	client influxdb.Client
}

var flagConfigFilename = flag.String("config.influxdb", ".config/influxdb.yaml", "InfluxDB config filename")

func (h *deviceHandler) GetName() string {
	return handlerName
}

func (h *deviceHandler) AddDevice(device *device.Device) {

}

func (h *deviceHandler) Start() error {
	// Load configuration
	reader, err := os.Open(*flagConfigFilename)
	if err != nil {
		// This is not fatal error, continue
		glog.Infof("Unable to load config: %v, influxDb disabled.", err)
		return nil
	}
	// Decode YAML
	decoder := yaml.NewDecoder(reader)
	decoder.SetStrict(true)
	if err := decoder.Decode(&h.config); err != nil {
		return err
	}

	if !h.config.Enabled {
		glog.Infof("InfluxDb disabled")
		return nil
	}

	h.client, err = influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:               h.config.Addr,
		Username:           h.config.Username,
		Password:           h.config.Password,
		InsecureSkipVerify: h.config.InsecureSkipVerify,
	})
	if err == nil {
		glog.Infof("Connected to %s", h.config.Addr)
	}

	return err
}

func (h *deviceHandler) Stop() {
	h.client.Close()

	if writer, err := os.Create(*flagConfigFilename); err == nil {
		encoder := yaml.NewEncoder(writer)
		encoder.Encode(&h.config)
	} else {
		glog.Infof("Unable to save config: %v", err)
	}
}

func (h *deviceHandler) ProcessMessage(device *device.Device, msg proto.Message) error {
	timestamp := time.Now()

	// Extract and log all proto field name/value pairs
	data := utils.ExtractAllNameValuesFromProtobuf(msg)

	// Common tags for all influxdb points
	tags := map[string]string{
		"device_id":    device.IDhex,
		"display_name": device.DisplayName,
	}

	// Group all values by table / metric
	measurements := map[string]kv{}
	for fullName, value := range data {
		tableName, metricName := splitProtobufFullName(fullName)
		if _, ok := measurements[tableName]; !ok {
			measurements[tableName] = map[string]interface{}{}
		}
		measurements[tableName][metricName] = value
	}

	// Write points into db
	points, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Precision: "s",
		Database:  h.config.Database,
	})
	if err != nil {
		return err
	}
	for tableName, values := range measurements {
		point, err := influxdb.NewPoint(tableName, tags, values, timestamp)
		if err != nil {
			return err
		}
		points.AddPoint(point)
	}

	return h.client.Write(points)
}

func splitProtobufFullName(fullName string) (string, string) {
	tokens := strings.SplitN(fullName, ".", 2)
	if len(tokens) > 1 {
		return tokens[0], tokens[1]
	}
	return "default", fullName
}

// Register device handler
func init() {
	device.MustAddHandler(&deviceHandler{})
}

package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
)

// ProtoGetFieldNameFromTag extracts original protobuf name from
// protobuf's field tag
func ProtoGetFieldNameFromTag(value string) string {
	tags := strings.Split(value, ",")
	for _, tag := range tags {
		kv := strings.Split(tag, "=")
		if len(kv) == 2 && kv[0] == "name" {
			return kv[1]
		}
	}
	return ""
}

// ExtractAllNameValuesFromProtobuf scans / extracts all protobuf fields / values into
// map of name -> value
func ExtractAllNameValuesFromProtobuf(msg proto.Message) map[string]interface{} {
	results := map[string]interface{}{}
	extractValuesRecursively("", reflect.ValueOf(msg), results)

	return results
}

func extractValuesRecursively(prefix string, value reflect.Value, results map[string]interface{}) {
	reflected := reflect.Indirect(value)
	for i := 0; i < reflected.NumField(); i++ {
		// skip protobuf internals (not marked with "protobuf" fields)
		tag, ok := reflected.Type().Field(i).Tag.Lookup("protobuf")
		if !ok {
			continue
		}
		fullName := getFullFieldName(prefix, ProtoGetFieldNameFromTag(tag))
		value := reflected.Field(i)
		// All structs in protobufs are pointers, process them recursively
		if value.Kind() == reflect.Ptr {
			// skip empty structures
			if value.IsNil() {
				continue
			}
			extractValuesRecursively(fullName, value, results)
		} else {
			// scalar types
			results[fullName] = value.Interface()
		}
	}
}

func getFullFieldName(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", prefix, name)
}

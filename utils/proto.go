package utils

import "strings"

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

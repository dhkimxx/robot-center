package domain

import (
	"encoding/json"
	"strings"
)

func InferSensorTypeFromID(sensorID string) string {
	normalized := strings.ToLower(strings.TrimSpace(sensorID))
	switch {
	case strings.Contains(normalized, "position"), strings.Contains(normalized, "location"), strings.Contains(normalized, "gps"):
		return "position"
	case strings.Contains(normalized, "imu"):
		return "imu"
	case strings.Contains(normalized, "gas"):
		return "gas"
	case strings.Contains(normalized, "point_cloud"), strings.Contains(normalized, "pointcloud"), strings.Contains(normalized, "lidar"):
		return "point_cloud"
	case strings.Contains(normalized, "event"), strings.Contains(normalized, "alarm"):
		return "event"
	default:
		return "unknown"
	}
}

func InferSensorValueType(sample SensorSample) string {
	if strings.TrimSpace(sample.ObjectKey) != "" {
		return "object_ref"
	}
	if len(sample.Values) == 0 {
		return "object"
	}
	var value any
	if err := json.Unmarshal(sample.Values, &value); err != nil {
		return "object"
	}
	switch value.(type) {
	case float64:
		return "number"
	case bool:
		return "boolean"
	case string:
		return "string"
	case []any:
		return "vector"
	default:
		return "object"
	}
}

package domain

import "strings"

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
	switch {
	case sample.NumericValue != nil:
		return "number"
	case sample.BoolValue != nil:
		return "boolean"
	case len(sample.VectorValue) > 0:
		return "vector"
	case strings.TrimSpace(sample.TextValue) != "":
		return "string"
	case strings.TrimSpace(sample.ObjectKey) != "":
		return "object_ref"
	default:
		return "object"
	}
}

package domain

import "strings"

func InferSensorTypeFromID(sensorID string) string {
	normalized := strings.ToLower(strings.TrimSpace(sensorID))
	switch {
	case strings.Contains(normalized, "position"), strings.Contains(normalized, "location"), strings.Contains(normalized, "gps"):
		return string(SensorTypePosition)
	case strings.Contains(normalized, "imu"):
		return string(SensorTypeIMU)
	case strings.Contains(normalized, "odometry"):
		return string(SensorTypeOdometry)
	case strings.Contains(normalized, "gas"):
		return string(SensorTypeGas)
	case strings.Contains(normalized, "point_cloud"), strings.Contains(normalized, "pointcloud"), strings.Contains(normalized, "lidar"):
		return string(SensorTypePointCloud)
	case strings.Contains(normalized, "battery"):
		return string(SensorTypeBattery)
	default:
		return string(SensorTypeUnknown)
	}
}

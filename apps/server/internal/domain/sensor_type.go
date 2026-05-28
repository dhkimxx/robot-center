package domain

import "strings"

type SensorType string

const (
	SensorTypeBattery     SensorType = "battery"
	SensorTypeGas         SensorType = "gas"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypeIMU         SensorType = "imu"
	SensorTypeOdometry    SensorType = "odometry"
	SensorTypePointCloud  SensorType = "point_cloud"
	SensorTypePosition    SensorType = "position"
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeUnknown     SensorType = "unknown"
)

func NormalizeSensorType(explicitType string, sensorID string) string {
	normalized := strings.ToLower(strings.TrimSpace(explicitType))
	if normalized == "" {
		return InferSensorTypeFromID(sensorID)
	}
	switch SensorType(normalized) {
	case SensorTypeBattery,
		SensorTypeGas,
		SensorTypeHumidity,
		SensorTypeIMU,
		SensorTypeOdometry,
		SensorTypePointCloud,
		SensorTypePosition,
		SensorTypeTemperature:
		return normalized
	default:
		return string(SensorTypeUnknown)
	}
}

func ParseSensorType(sensorType string) (SensorType, bool) {
	normalized := strings.ToLower(strings.TrimSpace(sensorType))
	switch SensorType(normalized) {
	case SensorTypeBattery,
		SensorTypeGas,
		SensorTypeHumidity,
		SensorTypeIMU,
		SensorTypeOdometry,
		SensorTypePointCloud,
		SensorTypePosition,
		SensorTypeTemperature:
		return SensorType(normalized), true
	default:
		return SensorTypeUnknown, false
	}
}

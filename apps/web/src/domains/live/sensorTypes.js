export const SensorType = Object.freeze({
  BATTERY: "battery",
  GAS: "gas",
  HUMIDITY: "humidity",
  IMU: "imu",
  ODOMETRY: "odometry",
  POINT_CLOUD: "point_cloud",
  POSITION: "position",
  TEMPERATURE: "temperature",
  UNKNOWN: "unknown"
});

export function normalizeSensorType(sensorType) {
  const normalized = String(sensorType ?? "").trim().toLowerCase();
  return Object.values(SensorType).includes(normalized) ? normalized : SensorType.UNKNOWN;
}

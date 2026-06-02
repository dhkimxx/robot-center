package domain

import (
	"encoding/json"
	"sort"
	"strings"
)

type SensorValueReading struct {
	FieldPath string  `json:"fieldPath"`
	Label     string  `json:"label"`
	Order     float64 `json:"order"`
	Unit      string  `json:"unit,omitempty"`
	Value     any     `json:"value"`
}

type sensorValueFieldDefinition struct {
	Label string
	Order float64
	Unit  string
}

var hiddenSensorValueFields = map[string]struct{}{
	"alarm":             {},
	"alarm_code":        {},
	"alarmCode":         {},
	"channel":           {},
	"frameId":           {},
	"high_alarm":        {},
	"highAlarm":         {},
	"latitude":          {},
	"longitude":         {},
	"low_alarm":         {},
	"lowAlarm":          {},
	"positionAvailable": {},
	"scale_code":        {},
	"scaleCode":         {},
	"valid":             {},
}

var sensorValueFieldDefinitions = map[string]sensorValueFieldDefinition{
	"accuracyMeter":       {Label: "위치 오차", Order: 46, Unit: "m"},
	"altitudeMeter":       {Label: "고도", Order: 43, Unit: "m"},
	"batteryPercent":      {Label: "배터리", Order: 30, Unit: "%"},
	"headingDegree":       {Label: "방향", Order: 44, Unit: "deg"},
	"networkState":        {Label: "네트워크", Order: 35},
	"speedMeterPerSecond": {Label: "속도", Order: 45, Unit: "m/s"},
	"yawDegree":           {Label: "Yaw", Order: 62, Unit: "deg"},
}

var sensorValueSegmentLabels = map[string]string{
	"angularVelocity":    "각속도",
	"linearAcceleration": "선가속도",
	"x":                  "X",
	"y":                  "Y",
	"z":                  "Z",
}

var sensorTypeReadingOrder = map[SensorType]float64{
	SensorTypeGas:        10,
	SensorTypeBattery:    30,
	SensorTypePosition:   40,
	SensorTypeIMU:        50,
	SensorTypeOdometry:   60,
	SensorTypePointCloud: 70,
	SensorTypeUnknown:    90,
}

var gasReadingLabelOrder = map[string]float64{
	"CO":   0.01,
	"H2S":  0.02,
	"O2":   0.03,
	"CH4":  0.04,
	"TEMP": 0.05,
	"HUM":  0.06,
}

func InterpretSensorSampleValue(descriptor SensorDescriptor, sample SensorSample) []SensorValueReading {
	value := decodeSensorSampleValue(sample.Values)
	if value == nil && sample.ObjectKey != "" {
		value = sample.ObjectKey
	}
	if value == nil {
		return []SensorValueReading{}
	}
	sensorType, _ := ParseSensorType(descriptor.SensorType)
	if sensorType == SensorTypeGas {
		return sortedSensorValueReadings(interpretGasSensorValue(descriptor, sample, value))
	}
	return sortedSensorValueReadings(interpretDefaultSensorValue(descriptor, sample, value))
}

func interpretGasSensorValue(descriptor SensorDescriptor, sample SensorSample, value any) []SensorValueReading {
	objectValue, ok := value.(map[string]any)
	if !ok {
		return interpretDefaultSensorValue(descriptor, sample, value)
	}
	concentration, ok := finiteNumber(objectValue["concentration"])
	if !ok {
		return interpretDefaultSensorValue(descriptor, sample, value)
	}
	unit := descriptor.Unit
	if valueUnit, ok := objectValue["unit"].(string); ok && strings.TrimSpace(valueUnit) != "" {
		unit = strings.TrimSpace(valueUnit)
	}
	return []SensorValueReading{{
		FieldPath: "concentration",
		Label:     nonEmptyString(descriptor.Label, descriptor.SensorID),
		Order:     sensorValueReadingOrder(descriptor, "concentration"),
		Unit:      unit,
		Value:     concentration,
	}}
}

func interpretDefaultSensorValue(descriptor SensorDescriptor, _ SensorSample, value any) []SensorValueReading {
	if objectValue, ok := value.(map[string]any); ok {
		readings := []SensorValueReading{}
		appendSensorObjectReadings(&readings, descriptor, "", objectValue)
		return readings
	}
	return []SensorValueReading{{
		FieldPath: descriptor.SensorID,
		Label:     nonEmptyString(descriptor.Label, descriptor.SensorID),
		Order:     sensorValueReadingOrder(descriptor, descriptor.SensorID),
		Unit:      descriptor.Unit,
		Value:     value,
	}}
}

func appendSensorObjectReadings(readings *[]SensorValueReading, descriptor SensorDescriptor, basePath string, objectValue map[string]any) {
	for fieldName, fieldValue := range objectValue {
		if _, hidden := hiddenSensorValueFields[fieldName]; hidden {
			continue
		}
		fieldPath := fieldName
		if basePath != "" {
			fieldPath = basePath + "." + fieldName
		}
		if nestedObject, ok := fieldValue.(map[string]any); ok {
			appendSensorObjectReadings(readings, descriptor, fieldPath, nestedObject)
			continue
		}
		if fieldValue == nil || fieldValue == "" {
			continue
		}
		*readings = append(*readings, SensorValueReading{
			FieldPath: fieldPath,
			Label:     makeSensorValueLabel(descriptor, fieldPath),
			Order:     sensorValueReadingOrder(descriptor, fieldPath),
			Unit:      sensorValueUnit(descriptor, fieldPath),
			Value:     fieldValue,
		})
	}
}

func makeSensorValueLabel(descriptor SensorDescriptor, fieldPath string) string {
	fieldName := lastPathSegment(fieldPath)
	if definition, ok := sensorValueFieldDefinitions[fieldName]; ok {
		return definition.Label
	}
	if fieldPath == descriptor.SensorID {
		return nonEmptyString(descriptor.Label, descriptor.SensorID)
	}
	segments := strings.Split(fieldPath, ".")
	for index, segment := range segments {
		if label, ok := sensorValueSegmentLabels[segment]; ok {
			segments[index] = label
		}
	}
	return strings.TrimSpace(nonEmptyString(descriptor.Label, descriptor.SensorID) + " " + strings.Join(segments, " "))
}

func sensorValueUnit(descriptor SensorDescriptor, fieldPath string) string {
	fieldName := lastPathSegment(fieldPath)
	if definition, ok := sensorValueFieldDefinitions[fieldName]; ok && strings.TrimSpace(definition.Unit) != "" {
		return definition.Unit
	}
	return descriptor.Unit
}

func sensorValueReadingOrder(descriptor SensorDescriptor, fieldPath string) float64 {
	fieldName := lastPathSegment(fieldPath)
	if definition, ok := sensorValueFieldDefinitions[fieldName]; ok {
		return definition.Order
	}
	sensorType, _ := ParseSensorType(descriptor.SensorType)
	order := sensorTypeReadingOrder[sensorType]
	if order == 0 {
		order = sensorTypeReadingOrder[SensorTypeUnknown]
	}
	if sensorType == SensorTypeGas {
		order += gasReadingLabelOrder[strings.ToUpper(strings.TrimSpace(descriptor.Label))]
	}
	return order
}

func sortedSensorValueReadings(readings []SensorValueReading) []SensorValueReading {
	sort.SliceStable(readings, func(leftIndex, rightIndex int) bool {
		left := readings[leftIndex]
		right := readings[rightIndex]
		if left.Order != right.Order {
			return left.Order < right.Order
		}
		if left.Label != right.Label {
			return left.Label < right.Label
		}
		return left.FieldPath < right.FieldPath
	})
	return readings
}

func decodeSensorSampleValue(rawValue json.RawMessage) any {
	if len(rawValue) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(rawValue, &value); err != nil {
		return nil
	}
	return value
}

func finiteNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	default:
		return 0, false
	}
}

func lastPathSegment(value string) string {
	segments := strings.Split(value, ".")
	return segments[len(segments)-1]
}

func nonEmptyString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}

package service

import (
	"context"
	"errors"
	"testing"
)

func TestSensorDataClearRequiresConfirmation(t *testing.T) {
	service := &SensorService{repository: &recordingStoreSpy{}}

	_, err := service.ClearSensorData(context.Background(), "development", "wrong")
	if !errors.Is(err, ErrSystemActionConfirmationRequired) {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestSensorDataClearIsDisabledInProduction(t *testing.T) {
	service := &SensorService{repository: &recordingStoreSpy{}}

	_, err := service.ClearSensorData(context.Background(), "production", clearSensorDataConfirmation)
	if !errors.Is(err, ErrSystemActionForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

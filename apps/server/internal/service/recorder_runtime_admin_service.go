package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

type RecorderRuntimeAdminService struct {
	environment               string
	recorderWorkerInternalURL string
	httpClient                *http.Client
}

func NewRecorderRuntimeAdminService(environment string, recorderWorkerInternalURL string) *RecorderRuntimeAdminService {
	return &RecorderRuntimeAdminService{
		environment:               environment,
		recorderWorkerInternalURL: strings.TrimRight(strings.TrimSpace(recorderWorkerInternalURL), "/"),
		httpClient:                &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *RecorderRuntimeAdminService) ClearRecorderRuntime(ctx context.Context, confirmation string) (domain.RecorderRuntimeClearResult, error) {
	if s == nil || strings.TrimSpace(s.recorderWorkerInternalURL) == "" {
		return domain.RecorderRuntimeClearResult{}, fmt.Errorf("recorder runtime admin service is not configured")
	}
	if strings.EqualFold(strings.TrimSpace(s.environment), "production") {
		return domain.RecorderRuntimeClearResult{}, ErrSystemActionForbidden
	}
	if strings.TrimSpace(confirmation) != domain.ClearRecorderRuntimeConfirmation {
		return domain.RecorderRuntimeClearResult{}, ErrSystemActionConfirmationRequired
	}

	requestBody, err := json.Marshal(map[string]string{"confirmation": domain.ClearRecorderRuntimeConfirmation})
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.recorderWorkerInternalURL+"/runtime/recordings/clear", bytes.NewReader(requestBody))
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		switch response.StatusCode {
		case http.StatusBadRequest:
			return domain.RecorderRuntimeClearResult{}, ErrSystemActionConfirmationRequired
		case http.StatusForbidden:
			return domain.RecorderRuntimeClearResult{}, ErrSystemActionForbidden
		case http.StatusConflict:
			return domain.RecorderRuntimeClearResult{}, fmt.Errorf("%w: %s", ErrSystemActionConflict, strings.TrimSpace(string(body)))
		default:
			return domain.RecorderRuntimeClearResult{}, fmt.Errorf("recorder-worker returned %s: %s", response.Status, strings.TrimSpace(string(body)))
		}
	}

	var payload struct {
		RecorderRuntime domain.RecorderRuntimeClearResult `json:"recorderRuntime"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.RecorderRuntimeClearResult{}, err
	}
	return payload.RecorderRuntime, nil
}

func (s *RecorderRuntimeAdminService) GetRecorderRuntimeStatus(ctx context.Context) (domain.RecorderRuntimeStatus, error) {
	if s == nil || strings.TrimSpace(s.recorderWorkerInternalURL) == "" {
		return domain.RecorderRuntimeStatus{}, fmt.Errorf("recorder runtime admin service is not configured")
	}
	statusContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(statusContext, http.MethodGet, s.recorderWorkerInternalURL+"/runtime/recordings/status", nil)
	if err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	response, err := s.httpClient.Do(request)
	if err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return domain.RecorderRuntimeStatus{}, fmt.Errorf("recorder-worker returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		RecorderRuntime domain.RecorderRuntimeStatus `json:"recorderRuntime"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.RecorderRuntimeStatus{}, err
	}
	return payload.RecorderRuntime, nil
}

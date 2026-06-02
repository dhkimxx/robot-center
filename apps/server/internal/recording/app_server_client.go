package recording

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
)

type AppServerClient interface {
	FetchRecordingTargets(ctx context.Context) ([]domain.Mission, error)
	CreateRecordingTick(ctx context.Context, target domain.Mission, chunkDuration time.Duration, tickAt time.Time) (domain.RecordingTickResult, error)
	ClaimRecordingFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error)
	MarkRecordingFinalizationJobCompleted(ctx context.Context, jobID string, uploadContext RecordingUploadContext) error
	MarkRecordingFinalizationJobPartial(ctx context.Context, jobID string, uploadContext RecordingUploadContext, reason string) error
	MarkRecordingFinalizationJobFailed(ctx context.Context, jobID string, uploadContext RecordingUploadContext, reason string) error
	MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, sizeBytes int64, uploadContext RecordingUploadContext) error
	MarkRecordingChunkUploaded(ctx context.Context, chunkID string, sizeBytes int64, uploadContext RecordingUploadContext) error
	PostDataChannelPayload(ctx context.Context, label string, payload []byte) error
}

type HTTPAppServerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPAppServerClient(baseURL string, httpClient *http.Client) *HTTPAppServerClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPAppServerClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *HTTPAppServerClient) FetchRecordingTargets(ctx context.Context) ([]domain.Mission, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(c.baseURL, "/api/v1/recorder/recording-targets"), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("app-server returned %s", response.Status)
	}

	var payload dto.RecordingTargetsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return missionsFromResponses(payload.Targets), nil
}

func (c *HTTPAppServerClient) CreateRecordingTick(ctx context.Context, target domain.Mission, chunkDuration time.Duration, tickAt time.Time) (domain.RecordingTickResult, error) {
	body := dto.RecorderTickRequest{
		MissionCode:          target.MissionCode,
		RobotCode:            target.RobotCode,
		ChunkDurationSeconds: int(chunkDuration.Seconds()),
		TickAt:               tickAt,
	}
	buffer, err := encodeJSONBuffer(body)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/v1/recorder/tick"), buffer)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return domain.RecordingTickResult{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return domain.RecordingTickResult{}, fmt.Errorf("app-server returned %s", response.Status)
	}

	var result dto.RecordingTickResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return domain.RecordingTickResult{}, err
	}
	return recordingTickResultFromResponse(result), nil
}

func (c *HTTPAppServerClient) ClaimRecordingFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error) {
	body := dto.RecorderFinalizationClaimRequest{
		WorkerID:            workerID,
		Limit:               limit,
		LockDurationSeconds: int(lockDuration.Seconds()),
	}
	buffer, err := encodeJSONBuffer(body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/v1/recorder/finalization-jobs/claim"), buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("app-server returned %s", response.Status)
	}
	var payload dto.RecorderFinalizationJobsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return recordingFinalizationJobsFromResponses(payload.Jobs), nil
}

func (c *HTTPAppServerClient) MarkRecordingFinalizationJobCompleted(ctx context.Context, jobID string, uploadContext RecordingUploadContext) error {
	return c.postRecordingFinalizationStatus(ctx, jobID, "completed", uploadContext, "")
}

func (c *HTTPAppServerClient) MarkRecordingFinalizationJobPartial(ctx context.Context, jobID string, uploadContext RecordingUploadContext, reason string) error {
	return c.postRecordingFinalizationStatus(ctx, jobID, "partial", uploadContext, reason)
}

func (c *HTTPAppServerClient) MarkRecordingFinalizationJobFailed(ctx context.Context, jobID string, uploadContext RecordingUploadContext, reason string) error {
	return c.postRecordingFinalizationStatus(ctx, jobID, "failed", uploadContext, reason)
}

func (c *HTTPAppServerClient) MarkRecordingFileUploaded(ctx context.Context, chunkID string, fileType string, sizeBytes int64, uploadContext RecordingUploadContext) error {
	request, err := c.newUploadNotificationRequest(ctx, joinURL(c.baseURL, "/api/v1/recorder/chunks/"+chunkID+"/files/"+fileType+"/uploaded"), sizeBytes, uploadContext)
	if err != nil {
		return err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("app-server returned %s", response.Status)
	}
	return nil
}

func (c *HTTPAppServerClient) MarkRecordingChunkUploaded(ctx context.Context, chunkID string, sizeBytes int64, uploadContext RecordingUploadContext) error {
	request, err := c.newUploadNotificationRequest(ctx, joinURL(c.baseURL, "/api/v1/recorder/chunks/"+chunkID+"/uploaded"), sizeBytes, uploadContext)
	if err != nil {
		return err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("app-server returned %s", response.Status)
	}
	return nil
}

func (c *HTTPAppServerClient) PostDataChannelPayload(ctx context.Context, label string, payload []byte) error {
	var path string
	switch label {
	case "channel.telemetry", "channel.spatial":
		path = "/api/v1/recorder/sensor-samples"
	default:
		return nil
	}
	if !json.Valid(payload) {
		return fmt.Errorf("invalid %s JSON payload", label)
	}

	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, joinURL(c.baseURL, path), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("app-server returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *HTTPAppServerClient) postRecordingFinalizationStatus(ctx context.Context, jobID string, status string, uploadContext RecordingUploadContext, reason string) error {
	body := dto.RecorderFinalizationStatusRequest{
		WorkerID: strings.TrimSpace(uploadContext.WorkerID),
		Attempt:  uploadContext.Attempt,
		Reason:   strings.TrimSpace(reason),
	}
	buffer, err := encodeJSONBuffer(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/v1/recorder/finalization-jobs/"+jobID+"/"+status), buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("app-server returned %s", response.Status)
	}
	return nil
}

func (c *HTTPAppServerClient) newUploadNotificationRequest(ctx context.Context, endpoint string, sizeBytes int64, uploadContext RecordingUploadContext) (*http.Request, error) {
	var requestSizeBytes *int64
	if sizeBytes > 0 {
		requestSizeBytes = &sizeBytes
	}
	body := dto.RecorderUploadRequest{
		SizeBytes: requestSizeBytes,
		WorkerID:  strings.TrimSpace(uploadContext.WorkerID),
		Attempt:   uploadContext.Attempt,
	}
	buffer, err := encodeJSONBuffer(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func encodeJSONBuffer(payload any) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(payload); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func joinURL(baseURL string, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

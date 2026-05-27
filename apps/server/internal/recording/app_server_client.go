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

type recordingTargetsResponse struct {
	Targets []domain.Mission `json:"targets"`
}

type recordingTickRequest struct {
	MissionCode          string    `json:"missionCode"`
	RobotCode            string    `json:"robotCode"`
	ChunkDurationSeconds int       `json:"chunkDurationSeconds"`
	TickAt               time.Time `json:"tickAt"`
}

type recordingFinalizationClaimRequest struct {
	WorkerID            string `json:"workerId"`
	Limit               int    `json:"limit"`
	LockDurationSeconds int    `json:"lockDurationSeconds"`
}

type recordingFinalizationJobsResponse struct {
	Jobs []domain.RecordingFinalizationJob `json:"jobs"`
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
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(c.baseURL, "/api/recording-targets"), nil)
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

	var payload recordingTargetsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Targets, nil
}

func (c *HTTPAppServerClient) CreateRecordingTick(ctx context.Context, target domain.Mission, chunkDuration time.Duration, tickAt time.Time) (domain.RecordingTickResult, error) {
	body := recordingTickRequest{
		MissionCode:          target.MissionCode,
		RobotCode:            target.RobotCode,
		ChunkDurationSeconds: int(chunkDuration.Seconds()),
		TickAt:               tickAt,
	}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return domain.RecordingTickResult{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/recorder/tick"), &buffer)
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

	var result domain.RecordingTickResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return domain.RecordingTickResult{}, err
	}
	return result, nil
}

func (c *HTTPAppServerClient) ClaimRecordingFinalizationJobs(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]domain.RecordingFinalizationJob, error) {
	body := recordingFinalizationClaimRequest{
		WorkerID:            workerID,
		Limit:               limit,
		LockDurationSeconds: int(lockDuration.Seconds()),
	}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/recorder/finalization-jobs/claim"), &buffer)
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
	var payload recordingFinalizationJobsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Jobs, nil
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
	request, err := c.newUploadNotificationRequest(ctx, joinURL(c.baseURL, "/api/recorder/chunks/"+chunkID+"/files/"+fileType+"/uploaded"), sizeBytes, uploadContext)
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
	request, err := c.newUploadNotificationRequest(ctx, joinURL(c.baseURL, "/api/recorder/chunks/"+chunkID+"/uploaded"), sizeBytes, uploadContext)
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
		path = "/api/sensor-samples"
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
	body := map[string]any{}
	if strings.TrimSpace(reason) != "" {
		body["reason"] = strings.TrimSpace(reason)
	}
	if strings.TrimSpace(uploadContext.WorkerID) != "" {
		body["workerId"] = strings.TrimSpace(uploadContext.WorkerID)
	}
	if uploadContext.Attempt > 0 {
		body["attempt"] = uploadContext.Attempt
	}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(c.baseURL, "/api/recorder/finalization-jobs/"+jobID+"/"+status), &buffer)
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
	body := map[string]any{}
	if sizeBytes > 0 {
		body["sizeBytes"] = sizeBytes
	}
	if strings.TrimSpace(uploadContext.WorkerID) != "" {
		body["workerId"] = strings.TrimSpace(uploadContext.WorkerID)
	}
	if uploadContext.Attempt > 0 {
		body["attempt"] = uploadContext.Attempt
	}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func joinURL(baseURL string, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

package domain

import "time"

const ClearRecorderRuntimeConfirmation = "CLEAR_RECORDER_RUNTIME"

type RecorderRuntimeClearResult struct {
	RecordingDirectoriesDeleted int   `json:"recordingDirectoriesDeleted"`
	FilesDeleted                int   `json:"filesDeleted"`
	DeletedBytes                int64 `json:"deletedBytes"`
}

type RecorderRuntimeStatus struct {
	Status               string    `json:"status"`
	RecordingDirectories int       `json:"recordingDirectories"`
	Files                int       `json:"files"`
	UsedBytes            int64     `json:"usedBytes"`
	TotalBytes           int64     `json:"totalBytes,omitempty"`
	AvailableBytes       int64     `json:"availableBytes,omitempty"`
	UsedPercent          float64   `json:"usedPercent,omitempty"`
	Clearable            bool      `json:"clearable"`
	BlockingReason       string    `json:"blockingReason,omitempty"`
	Error                string    `json:"error,omitempty"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

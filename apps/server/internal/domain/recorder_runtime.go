package domain

const ClearRecorderRuntimeConfirmation = "CLEAR_RECORDER_RUNTIME"

type RecorderRuntimeClearResult struct {
	RecordingDirectoriesDeleted int   `json:"recordingDirectoriesDeleted"`
	FilesDeleted                int   `json:"filesDeleted"`
	DeletedBytes                int64 `json:"deletedBytes"`
}

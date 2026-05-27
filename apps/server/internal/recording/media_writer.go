package recording

import (
	"context"
	"errors"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
	"robot-center/apps/server/internal/domain"
)

const minH264SnapshotBytes int64 = 64 * 1024
const minOggSnapshotBytes int64 = 4 * 1024
const minDataChannelSnapshotBytes int64 = 1

var errNoRecordingMedia = errors.New("no recording media available")

type activeAudioWriter struct {
	chunkID string
	path    string
	writer  *oggwriter.OggWriter
}

type h264ParameterSets struct {
	sps []byte
	pps []byte
}

type h264TrackTiming struct {
	haveTimestamp  bool
	firstTimestamp uint32
	lastTimestamp  uint32
	frameCount     int
}

type h264Snapshot struct {
	path string
	fps  float64
}

type recordingChunkFinalization struct {
	mediaKey string
	chunk    domain.RecordingChunk
}

type RecordingMediaUploadResult struct {
	UploadedFileTypes []string
}

type RecordingUploadContext struct {
	WorkerID string
	Attempt  int
}

type MediaUploader interface {
	UploadMediaSnapshots(ctx context.Context, roomID string, chunk domain.RecordingChunk, uploadContext RecordingUploadContext) (RecordingMediaUploadResult, error)
}

type mediaSnapshotter interface {
	createH264Snapshot(roomID string, chunkID string, label string) (h264Snapshot, error)
	createOggSnapshot(chunkID string) (string, error)
	createDataChannelSnapshot(chunkID string, label string) (string, error)
}

type recordingMediaUploader struct {
	appServerClient AppServerClient
	objectStorage   ObjectStorage
	snapshotter     mediaSnapshotter
}

func NewMediaUploader(appServerClient AppServerClient, objectStorage ObjectStorage, snapshotter mediaSnapshotter) MediaUploader {
	return &recordingMediaUploader{
		appServerClient: appServerClient,
		objectStorage:   objectStorage,
		snapshotter:     snapshotter,
	}
}

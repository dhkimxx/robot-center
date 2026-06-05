package recording

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"robot-center/apps/server/internal/domain"
	"strings"
)

func (u *recordingMediaUploader) UploadMediaSnapshots(ctx context.Context, roomID string, chunk domain.RecordingChunk, uploadContext RecordingUploadContext) (RecordingMediaUploadResult, error) {
	var uploadErrors []error
	result := RecordingMediaUploadResult{}
	if uploaded, err := u.muxAndUploadH264Snapshot(ctx, roomID, chunk, "rgb", "rgb_audio_mp4", chunk.MediaObjectKeys["rgbMp4"], uploadContext); err != nil {
		uploadErrors = append(uploadErrors, err)
	} else if uploaded {
		result.UploadedFileTypes = append(result.UploadedFileTypes, "rgb_audio_mp4")
	}
	if uploaded, err := u.muxAndUploadH264Snapshot(ctx, roomID, chunk, "thermal", "thermal_mp4", chunk.MediaObjectKeys["thermal"], uploadContext); err != nil {
		uploadErrors = append(uploadErrors, err)
	} else if uploaded {
		result.UploadedFileTypes = append(result.UploadedFileTypes, "thermal_mp4")
	}
	if uploaded, err := u.uploadDataChannelSnapshot(ctx, chunk, "sensor", "sensor_jsonl", chunk.MediaObjectKeys["sensor"], uploadContext); err != nil {
		uploadErrors = append(uploadErrors, err)
	} else if uploaded {
		result.UploadedFileTypes = append(result.UploadedFileTypes, "sensor_jsonl")
	}
	if uploaded, err := u.uploadDataChannelSnapshot(ctx, chunk, "telemetry", "telemetry_jsonl", chunk.MediaObjectKeys["telemetry"], uploadContext); err != nil {
		uploadErrors = append(uploadErrors, err)
	} else if uploaded {
		result.UploadedFileTypes = append(result.UploadedFileTypes, "telemetry_jsonl")
	}
	return result, errors.Join(uploadErrors...)
}

func (u *recordingMediaUploader) muxAndUploadH264Snapshot(ctx context.Context, roomID string, chunk domain.RecordingChunk, label string, fileType string, objectKey string, uploadContext RecordingUploadContext) (bool, error) {
	if strings.TrimSpace(objectKey) == "" {
		return false, nil
	}
	snapshot, err := u.snapshotter.createH264Snapshot(roomID, chunk.ID, label)
	if err != nil {
		return false, err
	}
	if snapshot.path == "" {
		return false, nil
	}
	defer os.Remove(snapshot.path)

	audioSnapshotPath := ""
	if label == "rgb" {
		audioSnapshotPath, err = u.snapshotter.createOggSnapshot(chunk.ID)
		if err != nil {
			log.Printf("recorder-worker audio snapshot skipped chunk=%s: %v", chunk.ID, err)
		}
		if audioSnapshotPath != "" {
			defer os.Remove(audioSnapshotPath)
		}
	}
	outputPath := filepath.Join(recordingChunkDirectory(chunk.ID), label+".mp4")
	if err := muxH264ToMP4(ctx, snapshot.path, audioSnapshotPath, outputPath, snapshot.fps, snapshot.durationSeconds); err != nil {
		return false, fmt.Errorf("%s mp4 mux failed: %w", label, err)
	}
	sizeBytes, err := u.objectStorage.UploadFile(ctx, objectKey, outputPath, "video/mp4")
	if err != nil {
		return false, fmt.Errorf("%s mp4 upload failed: %w", label, err)
	}
	if err := u.appServerClient.MarkRecordingFileUploaded(ctx, chunk.ID, fileType, sizeBytes, uploadContext); err != nil {
		return false, fmt.Errorf("%s file status update failed: %w", label, err)
	}
	log.Printf("recorder-worker mp4 uploaded chunk=%s label=%s key=%s", chunk.ID, label, objectKey)
	return true, nil
}

func (u *recordingMediaUploader) uploadDataChannelSnapshot(ctx context.Context, chunk domain.RecordingChunk, label string, fileType string, objectKey string, uploadContext RecordingUploadContext) (bool, error) {
	if strings.TrimSpace(objectKey) == "" {
		return false, nil
	}
	snapshotPath, err := u.snapshotter.createDataChannelSnapshot(chunk.ID, label)
	if err != nil {
		return false, err
	}
	if snapshotPath == "" {
		return false, nil
	}
	defer os.Remove(snapshotPath)

	sizeBytes, err := u.objectStorage.UploadFile(ctx, objectKey, snapshotPath, "application/x-ndjson")
	if err != nil {
		return false, fmt.Errorf("%s jsonl upload failed: %w", label, err)
	}
	if err := u.appServerClient.MarkRecordingFileUploaded(ctx, chunk.ID, fileType, sizeBytes, uploadContext); err != nil {
		return false, fmt.Errorf("%s file status update failed: %w", label, err)
	}
	log.Printf("recorder-worker jsonl uploaded chunk=%s label=%s key=%s", chunk.ID, label, objectKey)
	return true, nil
}

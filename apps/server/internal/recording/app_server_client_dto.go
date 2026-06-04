package recording

import (
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
)

func missionsFromResponses(responses []dto.RecorderRecordingTargetResponse) []domain.Mission {
	missions := make([]domain.Mission, 0, len(responses))
	for _, response := range responses {
		missions = append(missions, missionFromResponse(response))
	}
	return missions
}

func missionFromResponse(response dto.RecorderRecordingTargetResponse) domain.Mission {
	return domain.Mission{
		ID:          response.ID,
		MissionCode: response.MissionCode,
		Name:        response.Name,
		MissionType: response.MissionType,
		Status:      response.Status,
		SiteNote:    response.SiteNote,
		RobotCode:   response.RobotCode,
		RobotCodes:  append([]string(nil), response.RobotCodes...),
		StartedAt:   response.StartedAt,
		EndedAt:     response.EndedAt,
		CreatedAt:   response.CreatedAt,
		UpdatedAt:   response.UpdatedAt,
	}
}

func recordingTickResultFromResponse(response dto.RecorderRecordingTickResponse) domain.RecordingTickResult {
	return domain.RecordingTickResult{
		Chunk:    recordingChunkFromResponse(response.Chunk),
		Manifest: response.Manifest,
	}
}

func recordingChunkFromResponse(response dto.RecorderRecordingChunkResponse) domain.RecordingChunk {
	return domain.RecordingChunk{
		ID:                 response.ID,
		RecordingSessionID: response.RecordingSessionID,
		MissionID:          response.MissionID,
		MissionCode:        response.MissionCode,
		RobotCode:          response.RobotCode,
		ChunkIndex:         response.ChunkIndex,
		Status:             response.Status,
		StartedAt:          response.StartedAt,
		EndedAt:            response.EndedAt,
		DurationSeconds:    response.DurationSeconds,
		ManifestObjectKey:  response.ManifestObjectKey,
		MediaObjectKeys:    utils.CopyStringMap(response.MediaObjectKeys),
		AvailableFileTypes: utils.CopyBoolMap(response.AvailableFileTypes),
		CreatedAt:          response.CreatedAt,
		UpdatedAt:          response.UpdatedAt,
	}
}

func recordingFinalizationJobsFromResponses(responses []dto.RecorderFinalizationJobResponse) []domain.RecordingFinalizationJob {
	jobs := make([]domain.RecordingFinalizationJob, 0, len(responses))
	for _, response := range responses {
		jobs = append(jobs, recordingFinalizationJobFromResponse(response))
	}
	return jobs
}

func recordingFinalizationJobFromResponse(response dto.RecorderFinalizationJobResponse) domain.RecordingFinalizationJob {
	return domain.RecordingFinalizationJob{
		ID:                 response.ID,
		RecordingChunkID:   response.RecordingChunkID,
		RecordingSessionID: response.RecordingSessionID,
		MissionID:          response.MissionID,
		RobotID:            response.RobotID,
		Status:             response.Status,
		Reason:             response.Reason,
		Attempts:           response.Attempts,
		LockedBy:           response.LockedBy,
		LockedUntil:        response.LockedUntil,
		LastError:          response.LastError,
		CreatedAt:          response.CreatedAt,
		UpdatedAt:          response.UpdatedAt,
		CompletedAt:        response.CompletedAt,
		Chunk:              recordingChunkFromResponse(response.Chunk),
	}
}

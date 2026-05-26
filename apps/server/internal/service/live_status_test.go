package service

import (
	"testing"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
)

func TestLiveStatusRecordingIdleWhenChunkRecordingWithoutRecorderRuntime(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 50, 0, 0, time.UTC)
	lastTrackAt := now.Add(-2 * time.Second)
	service := LiveStatusService{}

	status := service.BuildMissionLiveStatus(LiveStatusInput{
		Mission: domain.Mission{
			ID:          "mission-id",
			MissionCode: "mission-007",
			Status:      "active",
			RobotCodes:  []string{"robot-001"},
		},
		Robots: []domain.Robot{{
			RobotCode:   "robot-001",
			DisplayName: "Robot 1",
			Status:      "online",
			LastSeenAt:  &now,
		}},
		ObservedRooms: []sfu.ObservedRoomSummary{{
			RoomID: "mission-007",
			Publishers: []sfu.ObservedPublisherSummary{{
				RobotCode:   "robot-001",
				ICEState:    "connected",
				TrackCount:  2,
				LastTrackAt: &lastTrackAt,
				UpdatedAt:   now,
			}},
		}},
		RecordingChunks: []domain.RecordingChunk{{
			ID:          "chunk-001",
			MissionCode: "mission-007",
			RobotCode:   "robot-001",
			Status:      "recording",
			StartedAt:   now.Add(-time.Minute),
			EndedAt:     now.Add(9 * time.Minute),
			UpdatedAt:   now.Add(-time.Second),
		}},
		Recorder:        RecorderRuntimeSnapshot{Available: true},
		Now:             now,
		FreshnessWindow: 30 * time.Second,
	})

	if got := status.Robots[0].Stream.State; got != "streaming" {
		t.Fatalf("stream state = %q, want streaming", got)
	}
	if got := status.Robots[0].Recording.State; got != "idle" {
		t.Fatalf("recording state = %q, want idle", got)
	}
	if got := status.Robots[0].Recording.Reason; got != "no_recorder_runtime" {
		t.Fatalf("recording reason = %q, want no_recorder_runtime", got)
	}
	if got := status.Robots[0].Recording.LatestChunkStatus; got != "recording" {
		t.Fatalf("latest chunk status = %q, want recording", got)
	}
}

func TestLiveStatusRecordingActiveWithFreshRecorderRuntime(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 50, 0, 0, time.UTC)
	lastTrackAt := now.Add(-2 * time.Second)
	service := LiveStatusService{}

	status := service.BuildMissionLiveStatus(LiveStatusInput{
		Mission: domain.Mission{
			MissionCode: "mission-007",
			Status:      "active",
			RobotCodes:  []string{"robot-001"},
		},
		ObservedRooms: []sfu.ObservedRoomSummary{{
			RoomID: "mission-007",
			Publishers: []sfu.ObservedPublisherSummary{{
				RobotCode:   "robot-001",
				ICEState:    "connected",
				TrackCount:  2,
				LastTrackAt: &lastTrackAt,
				UpdatedAt:   now,
			}},
		}},
		Recorder: RecorderRuntimeSnapshot{
			Available: true,
			Rooms: []RecorderRoomRuntime{{
				RoomID: "mission-007",
				Robots: []RecorderRobotRuntime{{
					RobotCode:   "robot-001",
					TrackCount:  2,
					LastTrackAt: &lastTrackAt,
				}},
			}},
		},
		Now:             now,
		FreshnessWindow: 30 * time.Second,
	})

	if got := status.Robots[0].Recording.State; got != "recording" {
		t.Fatalf("recording state = %q, want recording", got)
	}
}

func TestLiveStatusSeparatesRobotStreamState(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 50, 0, 0, time.UTC)
	lastTrackAt := now.Add(-2 * time.Second)
	service := LiveStatusService{}

	status := service.BuildMissionLiveStatus(LiveStatusInput{
		Mission: domain.Mission{
			MissionCode: "mission-007",
			Status:      "active",
			RobotCodes:  []string{"robot-001", "robot-002"},
		},
		ObservedRooms: []sfu.ObservedRoomSummary{{
			RoomID: "mission-007",
			Publishers: []sfu.ObservedPublisherSummary{{
				RobotCode:   "robot-001",
				ICEState:    "connected",
				TrackCount:  2,
				LastTrackAt: &lastTrackAt,
				UpdatedAt:   now,
			}},
		}},
		Now:             now,
		FreshnessWindow: 30 * time.Second,
	})

	if got := status.Robots[0].Stream.State; got != "streaming" {
		t.Fatalf("robot-001 stream = %q, want streaming", got)
	}
	if got := status.Robots[1].Stream.State; got != "waiting" {
		t.Fatalf("robot-002 stream = %q, want waiting", got)
	}
}

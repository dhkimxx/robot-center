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
			DeviceState: domain.RobotDeviceStateOnline,
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
	firstTrackAt := now.Add(-10 * time.Second)
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
				RobotCode:    "robot-001",
				ICEState:     "connected",
				TrackCount:   2,
				FirstTrackAt: &firstTrackAt,
				LastTrackAt:  &lastTrackAt,
				UpdatedAt:    now,
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
	if got := status.Robots[0].Stream.StartedAt; got == nil || !got.Equal(firstTrackAt) {
		t.Fatalf("stream startedAt = %#v, want %s", got, firstTrackAt)
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

func TestLiveStatusAppliesStreamSessionHistoryWithoutOverridingRuntimeState(t *testing.T) {
	now := time.Date(2026, 6, 4, 4, 35, 20, 0, time.UTC)
	lastMediaAt := now.Add(-time.Minute)
	previousEndedAt := now.Add(-2 * time.Minute)
	service := LiveStatusService{}

	status := service.BuildMissionLiveStatus(LiveStatusInput{
		Mission: domain.Mission{
			MissionCode: "mission-007",
			Status:      "active",
			RobotCodes:  []string{"robot-001"},
		},
		StreamSessions: []domain.RobotStreamSession{
			{
				MissionCode: "mission-007",
				RobotCode:   "robot-001",
				State:       "ended",
				StartedAt:   now.Add(-10 * time.Minute),
				LastMediaAt: &lastMediaAt,
				EndedAt:     &previousEndedAt,
				CreatedAt:   now.Add(-10 * time.Minute),
			},
			{
				MissionCode: "mission-007",
				RobotCode:   "robot-001",
				State:       "ended",
				StartedAt:   now.Add(-20 * time.Minute),
				CreatedAt:   now.Add(-20 * time.Minute),
			},
		},
		Now:             now,
		FreshnessWindow: 30 * time.Second,
	})

	stream := status.Robots[0].Stream
	if stream.State != "waiting" {
		t.Fatalf("stream state = %q, want waiting", stream.State)
	}
	if stream.LastMediaAt == nil || !stream.LastMediaAt.Equal(lastMediaAt) {
		t.Fatalf("lastMediaAt = %#v, want %s", stream.LastMediaAt, lastMediaAt)
	}
	if stream.PreviousEndedAt == nil || !stream.PreviousEndedAt.Equal(previousEndedAt) {
		t.Fatalf("previousEndedAt = %#v, want %s", stream.PreviousEndedAt, previousEndedAt)
	}
	if stream.ReconnectCount != 1 {
		t.Fatalf("reconnectCount = %d, want 1", stream.ReconnectCount)
	}
}

func TestLiveStatusConnectionUsesHeartbeatFreshness(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 50, 0, 0, time.UTC)
	staleSeenAt := now.Add(-2 * time.Minute)
	freshSeenAt := now.Add(-5 * time.Second)
	service := LiveStatusService{}

	status := service.BuildMissionLiveStatus(LiveStatusInput{
		Mission: domain.Mission{
			MissionCode: "mission-007",
			Status:      "active",
			RobotCodes:  []string{"robot-001", "robot-002", "robot-003"},
		},
		Robots: []domain.Robot{
			{RobotCode: "robot-001", DeviceState: domain.RobotDeviceStateOnline, LastSeenAt: &freshSeenAt},
			{RobotCode: "robot-002", DeviceState: domain.RobotDeviceStateOnline, LastSeenAt: &staleSeenAt},
			{RobotCode: "robot-003", DeviceState: domain.RobotDeviceStateFault, LastSeenAt: &freshSeenAt},
		},
		Now:             now,
		FreshnessWindow: 30 * time.Second,
	})

	if got := status.Robots[0].Connection.State; got != "online" {
		t.Fatalf("fresh robot connection = %q, want online", got)
	}
	if got := status.Robots[1].Connection.State; got != "disconnected" {
		t.Fatalf("stale robot connection = %q, want disconnected", got)
	}
	if got := status.Robots[2].Connection.State; got != "fault" {
		t.Fatalf("fault robot connection = %q, want fault", got)
	}
}

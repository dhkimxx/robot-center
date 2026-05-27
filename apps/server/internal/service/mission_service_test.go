package service

import (
	"context"
	"robot-center/apps/server/internal/domain"
	"testing"
)

func TestMissionServiceEndMissionQueuesRecordingFinalizationJobsInTransaction(t *testing.T) {
	outsideRepository := &recordingStoreSpy{}
	transactionRepository := &recordingStoreSpy{
		endMissionResult: domain.Mission{MissionCode: "mission-001", Status: "ended"},
	}
	transactionRunner := &recordingTransactionRunnerSpy{repository: transactionRepository}
	service := &MissionService{
		repository:          outsideRepository,
		recordingRepository: outsideRepository,
		transactionRunner:   transactionRunner,
	}

	mission, err := service.EndMission(context.Background(), "mission-001")
	if err != nil {
		t.Fatalf("EndMission returned error: %v", err)
	}
	if mission.MissionCode != "mission-001" || mission.Status != "ended" {
		t.Fatalf("mission = %#v", mission)
	}
	if !transactionRunner.called || !transactionRunner.committed {
		t.Fatalf("transaction runner called=%v committed=%v", transactionRunner.called, transactionRunner.committed)
	}
	if transactionRepository.endMissionInput != "mission-001" {
		t.Fatalf("transaction EndMission input = %q", transactionRepository.endMissionInput)
	}
	if transactionRepository.queuedFinalizationJobs != 1 {
		t.Fatalf("transaction queued finalization jobs = %d, want 1", transactionRepository.queuedFinalizationJobs)
	}
	if outsideRepository.endMissionInput != "" || outsideRepository.queuedFinalizationJobs != 0 {
		t.Fatalf("outside repository was used: end=%q queued=%d", outsideRepository.endMissionInput, outsideRepository.queuedFinalizationJobs)
	}
}

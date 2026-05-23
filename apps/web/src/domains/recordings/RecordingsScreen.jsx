import { useEffect, useMemo, useState } from "react";
import { formatDateTime, formatDurationSeconds, makeStatusLabel } from "../../utils/formatters.js";
import {
  createRecordingPlaybackFile,
  getRecordingObjectEntries,
  isPlayableRecordingFile,
  makeFileStatusLabel,
  makeRecordingRobotGroups,
  makeRecordingSessionGroups
} from "./recordingHelpers.js";

export default function RecordingsScreen({ onOpenPlaybackFile, recordings }) {
  const sessionGroups = useMemo(() => makeRecordingSessionGroups(recordings), [recordings]);
  const robotGroups = useMemo(() => makeRecordingRobotGroups(sessionGroups), [sessionGroups]);
  const [selectedRobotCode, setSelectedRobotCode] = useState("");
  const selectedRobotGroup = robotGroups.find((group) => group.robotCode === selectedRobotCode) ?? robotGroups[0] ?? null;

  useEffect(() => {
    if (robotGroups.length === 0) {
      setSelectedRobotCode("");
      return;
    }
    if (!selectedRobotCode || !robotGroups.some((group) => group.robotCode === selectedRobotCode)) {
      setSelectedRobotCode(robotGroups[0].robotCode);
    }
  }, [robotGroups, selectedRobotCode]);

  return (
    <article className="surface recording-page">
      <div className="section-heading">
        <h2>로봇별 녹화</h2>
        <span>{robotGroups.length}대 / {sessionGroups.length}개 세션 / {recordings.length}개 청크</span>
      </div>
      {robotGroups.length === 0 ? (
        <div className="recording-session-list">
          <p className="empty-state">아직 생성된 녹화 메타데이터가 없습니다.</p>
        </div>
      ) : (
        <div className="recording-browser">
          <aside className="recording-robot-list">
            {robotGroups.map((group) => (
              <button
                className={selectedRobotGroup?.robotCode === group.robotCode ? "recording-robot active" : "recording-robot"}
                key={group.robotCode}
                type="button"
                onClick={() => setSelectedRobotCode(group.robotCode)}
              >
                <strong>{group.robotCode}</strong>
                <span>{group.sessionCount}개 세션 / {group.chunkCount}개 청크</span>
                <small>최근 {formatDateTime(group.latestAt)}</small>
              </button>
            ))}
          </aside>
          <section className="recording-session-list">
            <div className="recording-list-header">
              <div>
                <strong>{selectedRobotGroup?.robotCode}</strong>
                <span>최근 녹화순</span>
              </div>
              <small>{selectedRobotGroup?.sessionCount ?? 0}개 세션</small>
            </div>
            {(selectedRobotGroup?.sessions ?? []).map((session) => (
              <section className="recording-session-card" key={session.id}>
                <div className="recording-session-header">
                  <div>
                    <strong>{session.missionCode}</strong>
                    <span>{formatDateTime(session.startedAt)} - {formatDateTime(session.endedAt)}</span>
                  </div>
                  <div className="recording-session-summary">
                    <span className={`file-status ${session.status === "uploaded" ? "available" : "recording"}`}>{makeStatusLabel(session.status)}</span>
                    <small>{session.availableFileCount}/{session.fileCount} 파일 저장</small>
                  </div>
                </div>
                <div className="recording-chunk-list">
                  {session.chunks.map((recording) => (
                    <div className="recording-row" key={recording.id}>
                      <div className="recording-row-main">
                        <strong>청크 #{recording.chunkIndex}</strong>
                        <span>{formatDurationSeconds(recording.durationSeconds)} / {formatDateTime(recording.startedAt)} - {formatDateTime(recording.endedAt)}</span>
                        <span>상태 {makeStatusLabel(recording.status)} / 갱신 {formatDateTime(recording.updatedAt)}</span>
                      </div>
                      <RecordingObjectList
                        onSelectVideo={(entry) => onOpenPlaybackFile(createRecordingPlaybackFile(recording, entry))}
                        recording={recording}
                      />
                    </div>
                  ))}
                </div>
              </section>
            ))}
          </section>
        </div>
      )}
    </article>
  );
}

function RecordingObjectList({ onSelectVideo, recording }) {
  const entries = getRecordingObjectEntries(recording);
  return (
    <div className="object-list">
      {entries.map((entry) => (
        <div className="object-row" key={`${recording.id}-${entry.type ?? entry.label}`}>
          <div>
            <strong>{entry.label}</strong>
            <span>{entry.contentType ?? "메타데이터"}</span>
          </div>
          {isPlayableRecordingFile(entry) ? (
            <button className="object-link" type="button" onClick={() => onSelectVideo?.(entry)}>
              재생
            </button>
          ) : entry.status === "available" && entry.url ? (
            <a className="object-link" href={entry.url} target="_blank" rel="noreferrer">
              {entry.type === "manifest" ? "보기" : "열기"}
            </a>
          ) : (
            <span className={`file-status ${entry.status}`}>{makeFileStatusLabel(entry.status)}</span>
          )}
        </div>
      ))}
    </div>
  );
}

export function RecordingPlaybackModal({ file, onClose }) {
  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  return (
    <div className="playback-modal-backdrop" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget) {
        onClose();
      }
    }}>
      <section className="playback-modal" role="dialog" aria-modal="true" aria-label="녹화 플레이어">
        <header className="playback-modal-header">
          <div>
            <strong>{file.label}</strong>
            <span>{file.missionCode} / {file.robotCode} / 청크 #{file.chunkIndex}</span>
          </div>
          <div className="playback-modal-actions">
            <small>{formatDateTime(file.startedAt)} - {formatDateTime(file.endedAt)}</small>
            <button className="secondary-button" type="button" onClick={onClose}>닫기</button>
          </div>
        </header>
        <div className="playback-modal-body">
          <video key={file.url} src={file.url} controls playsInline />
        </div>
      </section>
    </div>
  );
}

import {
  formatDateTime,
  makeStatusLabel
} from "../../../utils/formatters.js";
import { createRecordingPlaybackFile } from "../../recordings/recordingHelpers.js";

export function LatestRecordingPanel({ recording, compact = false, onOpenRecordings, onPlayRecording, playbackRecording }) {
  const playableFile = createRecordingPlaybackFile(playbackRecording ?? recording);
  return (
    <article className={compact ? "latest-recording compact" : "surface latest-recording"}>
      <div className="section-heading">
        <h2>녹화 저장</h2>
        <span>{playableFile ? "재생 가능" : recording ? makeStatusLabel(recording.status) : "대기"}</span>
      </div>
      {!recording ? (
        <p className="empty-state">진행 중인 임무가 시작되면 저장 청크가 생성됩니다.</p>
      ) : (
        <div className="latest-recording-body">
          <div>
            <strong>{recording.missionCode} / #{recording.chunkIndex}</strong>
            <span>{formatDateTime(recording.startedAt)} - {formatDateTime(recording.endedAt)}</span>
          </div>
          <div className="latest-recording-actions">
            {playableFile ? (
              <button className="object-link" type="button" onClick={onPlayRecording}>
                재생
              </button>
            ) : null}
            <button className="object-link secondary" type="button" onClick={onOpenRecordings}>
              녹화 목록
            </button>
          </div>
        </div>
      )}
    </article>
  );
}

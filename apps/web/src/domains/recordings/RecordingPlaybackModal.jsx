import { useEffect } from "react";
import { formatDateTime } from "../../utils/formatters.js";

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

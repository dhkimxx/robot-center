import Button from "../../components/ui/Button.jsx";
import { useDialogLifecycle } from "../../hooks/useDialogLifecycle.js";
import { formatDateTime } from "../../utils/formatters.js";

export function RecordingPlaybackModal({ file, onClose }) {
  const dialogRef = useDialogLifecycle(onClose);

  return (
    <div className="fixed inset-0 z-[20000] grid place-items-center bg-command-950/55 p-6 backdrop-blur-sm max-[820px]:p-3" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget) {
        onClose();
      }
    }}>
      <section
        className="grid max-h-[calc(100vh-48px)] w-[min(1100px,calc(100vw-48px))] grid-rows-[auto_minmax(0,1fr)] overflow-hidden rounded-2xl border border-slate-500/20 bg-command-950 shadow-command max-[820px]:max-h-[calc(100vh-24px)] max-[820px]:w-[calc(100vw-24px)]"
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label="녹화 플레이어"
        tabIndex={-1}
      >
        <header className="flex items-center justify-between gap-4 border-b border-slate-500/15 bg-command-900 px-5 py-4 max-[820px]:flex-col max-[820px]:items-stretch">
          <div className="min-w-0">
            <strong className="block truncate text-base font-bold text-slate-50">{file.label}</strong>
            <span className="mt-1 block text-sm font-semibold text-slate-400">{file.missionCode} / {file.robotCode} / 청크 #{file.chunkIndex}</span>
          </div>
          <div className="flex shrink-0 items-center gap-2 max-[820px]:flex-col max-[820px]:items-stretch">
            <small className="text-xs font-bold text-slate-400">{formatDateTime(file.startedAt)} - {formatDateTime(file.endedAt)}</small>
            <Button onClick={onClose}>닫기</Button>
          </div>
        </header>
        <div className="grid min-h-[360px] bg-command-950">
          <video className="block h-[min(72vh,760px)] w-full bg-command-950 object-contain" key={file.url} src={file.url} controls playsInline />
        </div>
      </section>
    </div>
  );
}

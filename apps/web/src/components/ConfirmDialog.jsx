import { RiCloseLine } from "react-icons/ri";
import { useDialogLifecycle } from "../hooks/useDialogLifecycle.js";
import { cn } from "../utils/cn.js";
import Button from "./ui/Button.jsx";

export default function ConfirmDialog({
  cancelLabel = "취소",
  confirmLabel = "확인",
  description,
  onCancel,
  onConfirm,
  subject,
  title,
  tone = "default"
}) {
  const dialogRef = useDialogLifecycle(onCancel);

  return (
    <div
      className="fixed inset-0 z-[21000] grid place-items-center bg-command-950/35 p-5 backdrop-blur-[2px]"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onCancel();
        }
      }}
    >
      <section
        aria-label={title}
        aria-modal="true"
        className={cn(
          "grid w-[min(420px,calc(100vw-40px))] overflow-hidden rounded-2xl border border-slate-500/20 bg-command-800 p-0 shadow-command focus:outline-none",
          tone === "danger" && "border-red-400/25"
        )}
        ref={dialogRef}
        role="alertdialog"
        tabIndex={-1}
      >
        <header className="flex items-center justify-between gap-3 border-b border-slate-500/15 px-5 py-4">
          <div className="grid min-w-0 flex-1 gap-1">
            <strong className="text-base font-bold leading-tight text-slate-50">{title}</strong>
            {subject ? <span className="truncate text-xs font-semibold text-slate-400">{subject}</span> : null}
          </div>
          <button
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-300 transition hover:bg-white/[0.06] hover:text-white"
            type="button"
            aria-label="닫기"
            onClick={onCancel}
          >
            <RiCloseLine aria-hidden="true" />
          </button>
        </header>
        {description ? <p className="px-5 pt-4 text-sm font-medium leading-relaxed text-slate-400">{description}</p> : null}
        <footer className="flex justify-end gap-2 px-5 py-5">
          <Button onClick={onCancel}>{cancelLabel}</Button>
          <Button variant={tone === "danger" ? "danger" : "primary"} onClick={() => void onConfirm()}>
            {confirmLabel}
          </Button>
        </footer>
      </section>
    </div>
  );
}

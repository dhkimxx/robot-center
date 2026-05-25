import { RiCloseLine } from "react-icons/ri";
import { useDialogLifecycle } from "../hooks/useDialogLifecycle.js";
import { cn } from "../utils/cn.js";

const modalSizeClasses = {
  large: "w-[min(680px,calc(100vw-56px))]",
  medium: "w-[min(520px,calc(100vw-48px))]"
};

export default function Modal({ children, description, footer, onClose, size = "medium", title }) {
  const dialogRef = useDialogLifecycle(onClose);

  return (
    <div
      className="fixed inset-0 z-[20000] grid place-items-center bg-command-950/55 p-6 backdrop-blur-sm max-[820px]:p-3"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <section
        aria-label={title}
        aria-modal="true"
        className={cn(
          "grid max-h-[calc(100vh-48px)] overflow-hidden rounded-2xl border border-slate-500/20 bg-command-800 shadow-command focus:outline-none max-[820px]:w-[calc(100vw-24px)] max-[820px]:max-h-[calc(100vh-24px)]",
          "grid-rows-[auto_minmax(0,1fr)_auto]",
          modalSizeClasses[size] ?? modalSizeClasses.medium
        )}
        ref={dialogRef}
        role="dialog"
        tabIndex={-1}
      >
        <header className="flex items-center justify-between gap-3 border-b border-slate-500/15 bg-command-900 px-5 py-4 max-[820px]:items-stretch">
          <div className="min-w-0">
            <strong className="block truncate text-base font-bold text-slate-50">{title}</strong>
            {description ? <span className="mt-1 block text-sm font-semibold text-slate-400">{description}</span> : null}
          </div>
          <button
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-300 transition hover:bg-white/[0.06] hover:text-white"
            type="button"
            aria-label="닫기"
            onClick={onClose}
          >
            <RiCloseLine aria-hidden="true" />
          </button>
        </header>
        <div className="min-h-0 overflow-auto bg-command-800 p-5 text-slate-100">
          {children}
        </div>
        {footer ? <footer className="flex justify-end gap-2 border-t border-slate-500/15 bg-command-900 px-5 py-4 max-[820px]:flex-col-reverse">{footer}</footer> : null}
      </section>
    </div>
  );
}

import { useCallback, useEffect, useRef, useState } from "react";
import { RiCloseLine } from "react-icons/ri";
import { cn } from "../utils/cn.js";

const notificationAutoDismissMs = 4000;
const notificationStaggerMs = 450;
const notificationExitMs = 180;

const notificationLabels = {
  danger: "오류",
  info: "알림",
  success: "완료",
  warning: "확인"
};

const notificationToneClasses = {
  danger: "bg-red-950/95 text-red-100",
  info: "bg-command-800/95 text-slate-100",
  success: "bg-emerald-950/95 text-emerald-100",
  warning: "bg-amber-950/95 text-amber-100"
};

function NotificationItem({ index, notification, onDismiss }) {
  const [isClosing, setIsClosing] = useState(false);
  const dismissTimerRef = useRef(null);
  const removeTimerRef = useRef(null);

  const startDismiss = useCallback(() => {
    setIsClosing((current) => {
      if (current) {
        return current;
      }
      removeTimerRef.current = window.setTimeout(() => {
        onDismiss(notification.id);
      }, notificationExitMs);
      return true;
    });
  }, [notification.id, onDismiss]);

  useEffect(() => {
    dismissTimerRef.current = window.setTimeout(startDismiss, notificationAutoDismissMs + index * notificationStaggerMs);
    return () => {
      window.clearTimeout(dismissTimerRef.current);
      window.clearTimeout(removeTimerRef.current);
    };
  }, [notification.id, startDismiss]);

  const tone = notification.tone ?? "info";

  return (
    <div
      className={cn(
        "grid min-h-12 grid-cols-[64px_minmax(0,1fr)_28px] items-center gap-3 rounded-xl border border-slate-500/20 px-3 py-2 shadow-command transition duration-200",
        notificationToneClasses[tone] ?? notificationToneClasses.info,
        isClosing ? "-translate-y-2 opacity-0" : "translate-y-0 opacity-100"
      )}
      role={tone === "danger" ? "alert" : "status"}
    >
      <strong className="text-xs font-black">{notificationLabels[tone] ?? notificationLabels.info}</strong>
      <span className="min-w-0 truncate text-sm font-semibold">{notification.message}</span>
      <button
        className="inline-flex h-7 w-7 items-center justify-center rounded-lg text-inherit transition hover:bg-white/[0.08]"
        type="button"
        aria-label="알림 닫기"
        onClick={startDismiss}
      >
        <RiCloseLine aria-hidden="true" />
      </button>
    </div>
  );
}

export default function NotificationStack({ notifications, onDismiss }) {
  if (notifications.length === 0) {
    return null;
  }

  return (
    <div className="fixed right-4 top-4 z-[13000] grid w-[min(460px,calc(100vw-32px))] gap-2 max-[820px]:left-3 max-[820px]:right-3 max-[820px]:top-3 max-[820px]:w-auto" aria-live="polite">
      {notifications.map((notification, index) => (
        <NotificationItem
          index={index}
          key={notification.id}
          notification={notification}
          onDismiss={onDismiss}
        />
      ))}
    </div>
  );
}

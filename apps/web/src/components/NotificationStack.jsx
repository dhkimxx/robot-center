import { useCallback, useEffect, useRef, useState } from "react";

const notificationAutoDismissMs = 4000;
const notificationStaggerMs = 450;
const notificationExitMs = 180;

const notificationLabels = {
  danger: "오류",
  info: "알림",
  success: "완료",
  warning: "확인"
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
    <div className={`notification-banner ${tone} ${isClosing ? "closing" : ""}`} role={tone === "danger" ? "alert" : "status"}>
      <strong>{notificationLabels[tone] ?? notificationLabels.info}</strong>
      <span>{notification.message}</span>
      <button className="notification-close-button" type="button" aria-label="알림 닫기" onClick={startDismiss}>X</button>
    </div>
  );
}

export default function NotificationStack({ notifications, onDismiss }) {
  if (notifications.length === 0) {
    return null;
  }

  return (
    <div className="notification-region" aria-live="polite">
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

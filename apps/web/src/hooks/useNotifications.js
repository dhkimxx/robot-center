import { useCallback, useRef, useState } from "react";

export function useNotifications() {
  const [notifications, setNotifications] = useState([]);
  const notificationSequenceRef = useRef(0);

  const showNotification = useCallback((message, tone = "info") => {
    notificationSequenceRef.current += 1;
    const notification = {
      id: `notification-${Date.now()}-${notificationSequenceRef.current}`,
      message,
      tone
    };
    setNotifications((current) => [...current, notification].slice(-5));
  }, []);

  const dismissNotification = useCallback((notificationId) => {
    setNotifications((current) => current.filter((notification) => notification.id !== notificationId));
  }, []);

  return {
    notifications,
    showNotification,
    dismissNotification
  };
}

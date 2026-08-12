// useNotificationStatus.ts
// Polls a notification's status until it reaches a terminal state (sent/failed/cancelled).
// Useful for showing a live "sending... -> sent" indicator in a dashboard,
// e.g. after triggering an AuraMed appointment reminder.

import { useEffect, useRef, useState } from "react";
import { heraldClient, Notification, NotificationStatus } from "../heraldClient";

const TERMINAL_STATUSES: NotificationStatus[] = ["sent", "failed", "cancelled"];

export function useNotificationStatus(notificationId: string | null, pollIntervalMs = 2000) {
  const [notification, setNotification] = useState<Notification | null>(null);
  const [error, setError] = useState<string | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!notificationId) return;

    const poll = async () => {
      try {
        const result = await heraldClient.getNotification(notificationId);
        setNotification(result);

        if (TERMINAL_STATUSES.includes(result.status) && intervalRef.current) {
          clearInterval(intervalRef.current);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch notification");
        if (intervalRef.current) clearInterval(intervalRef.current);
      }
    };

    poll(); // fire immediately, then on interval
    intervalRef.current = setInterval(poll, pollIntervalMs);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [notificationId, pollIntervalMs]);

  return { notification, error, isTerminal: notification ? TERMINAL_STATUSES.includes(notification.status) : false };
}

// NotificationSender.example.tsx
// Example usage: a form that sends a notification via Herald and shows live status.
// Drop into any React admin dashboard (AuraMed staff panel, StoreCore admin, etc.)

import { useState } from "react";
import { heraldClient } from "./heraldClient";
import { useNotificationStatus } from "./hooks/useNotificationStatus";

export function NotificationSender() {
  const [recipient, setRecipient] = useState("");
  const [body, setBody] = useState("");
  const [notificationId, setNotificationId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const { notification, error, isTerminal } = useNotificationStatus(notificationId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const result = await heraldClient.createNotification({
        channel: "email",
        recipient,
        subject: "AuraMed Appointment Reminder",
        body,
        priority: "high",
        metadata: { source: "admin-dashboard" },
      });
      setNotificationId(result.id);
    } catch (err) {
      console.error("Failed to send notification:", err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <input
          type="email"
          placeholder="Recipient email"
          value={recipient}
          onChange={(e) => setRecipient(e.target.value)}
          required
        />
        <textarea
          placeholder="Message body"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          required
        />
        <button type="submit" disabled={submitting}>
          {submitting ? "Sending..." : "Send Notification"}
        </button>
      </form>

      {notification && (
        <div>
          <p>Status: <strong>{notification.status}</strong></p>
          <p>Attempts: {notification.attempt_count} / {notification.max_attempts}</p>
          {isTerminal && notification.status === "sent" && <p>✅ Delivered</p>}
          {isTerminal && notification.status === "failed" && <p>❌ Delivery failed</p>}
        </div>
      )}

      {error && <p style={{ color: "red" }}>{error}</p>}
    </div>
  );
}

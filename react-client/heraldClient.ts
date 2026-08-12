// heraldClient.ts
// Thin typed wrapper around Herald's REST API. Drop this into any React app
// (admin dashboard for AuraMed/StoreCore/LMS) and use it as-is.

const HERALD_BASE_URL = import.meta.env.VITE_HERALD_API_URL ?? "http://localhost:8080/api/v1";
const HERALD_API_KEY = import.meta.env.VITE_HERALD_API_KEY; // set per-tenant in .env

export type NotificationChannel = "email" | "sms" | "push";
export type NotificationStatus =
  | "pending" | "queued" | "sending" | "sent" | "failed" | "retrying" | "cancelled";

export interface Notification {
  id: string;
  tenant_id: string;
  channel: NotificationChannel;
  status: NotificationStatus;
  priority: string;
  recipient: string;
  subject?: string;
  body: string;
  metadata: Record<string, unknown>;
  attempt_count: number;
  max_attempts: number;
  created_at: string;
  sent_at?: string;
  failed_at?: string;
}

export interface CreateNotificationInput {
  channel: NotificationChannel;
  recipient: string;
  subject?: string;
  body: string;
  priority?: "low" | "normal" | "high" | "urgent";
  metadata?: Record<string, unknown>;
}

interface ApiSuccess<T> {
  success: true;
  data: T;
}
interface ApiError {
  success: false;
  error: string;
  code?: string;
}

class HeraldError extends Error {
  code?: string;
  constructor(message: string, code?: string) {
    super(message);
    this.code = code;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${HERALD_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${HERALD_API_KEY}`,
      ...options.headers,
    },
  });

  const json = (await res.json()) as ApiSuccess<T> | ApiError;

  if (!json.success) {
    throw new HeraldError(json.error, json.code);
  }

  return json.data;
}

export const heraldClient = {
  createNotification: (input: CreateNotificationInput) =>
    request<Notification>("/notifications", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  getNotification: (id: string) =>
    request<Notification>(`/notifications/${id}`),

  listNotifications: (params?: { status?: NotificationStatus; limit?: number; offset?: number }) => {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.limit) query.set("limit", String(params.limit));
    if (params?.offset) query.set("offset", String(params.offset));
    const qs = query.toString();
    return request<Notification[]>(`/notifications${qs ? `?${qs}` : ""}`);
  },
};

export { HeraldError };

import axios, { type AxiosInstance, type AxiosResponse } from 'axios';

// ── Base client ──────────────────────────────────────────────────────────────

export const apiClient: AxiosInstance = axios.create({
  baseURL: 'https://syncguard-adminsvc.fly.dev',
  timeout: 15_000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach bearer token if present in localStorage
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('sg_admin_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Surface error details for callers
apiClient.interceptors.response.use(
  (res) => res,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      localStorage.removeItem('sg_admin_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  },
);

// ── Shared primitives ────────────────────────────────────────────────────────

export type DeviceStatus = 'approved' | 'pending' | 'blocked' | 'quarantined';
export type UserStatus = 'active' | 'disabled' | 'locked';
export type UserSource = 'ldap' | 'local' | 'entra';
export type AuditSeverity = 'info' | 'warning' | 'critical';
export type AuditResult = 'allow' | 'deny' | 'error';

// ── Device types ─────────────────────────────────────────────────────────────

export interface Device {
  id: string;
  deviceId: string;
  userId: string;
  username: string;
  type: string;
  os: string;
  status: DeviceStatus;
  lastSeen: string; // ISO-8601
  enrolledAt: string;
  userAgent: string;
}

export interface DeviceListResponse {
  items: Device[];
  total: number;
  page: number;
  pageSize: number;
}

export interface DeviceActionRequest {
  action: 'approve' | 'block' | 'quarantine';
  reason?: string;
}

// ── User types ───────────────────────────────────────────────────────────────

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  displayName: string;
  status: UserStatus;
  source: UserSource;
  mfaEnrolled: boolean;
  deviceCount: number;
  lastLogin: string | null;
  createdAt: string;
  groups: string[];
}

export interface UserListResponse {
  items: AdminUser[];
  total: number;
  page: number;
  pageSize: number;
}

// ── Audit log types ──────────────────────────────────────────────────────────

export interface AuditEvent {
  id: string;
  timestamp: string;
  action: string;
  username: string;
  deviceId: string | null;
  ipAddress: string;
  result: AuditResult;
  severity: AuditSeverity;
  detail: string;
  policyId: string | null;
}

export interface AuditListResponse {
  items: AuditEvent[];
  total: number;
  page: number;
  pageSize: number;
}

// ── Dashboard / stats types ──────────────────────────────────────────────────

export interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  totalDevices: number;
  pendingApprovals: number;
  blockedDevices: number;
  authEventsLast24h: number;
  mfaFailuresLast24h: number;
}

// ── Policy types ─────────────────────────────────────────────────────────────

export type PolicyEffect = 'allow' | 'deny';
export type PolicyConditionType = 'device_status' | 'user_group' | 'ip_range' | 'time_window';

export interface PolicyCondition {
  type: PolicyConditionType;
  operator: 'eq' | 'in' | 'not_in' | 'cidr';
  value: string | string[];
}

export interface Policy {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  effect: PolicyEffect;
  conditions: PolicyCondition[];
  createdAt: string;
  updatedAt: string;
}

export interface PolicyListResponse {
  items: Policy[];
  total: number;
}

// ── API call helpers ─────────────────────────────────────────────────────────

function unwrap<T>(res: AxiosResponse<T>): T {
  return res.data;
}

export const devicesApi = {
  list: (page = 1, pageSize = 50): Promise<DeviceListResponse> =>
    apiClient.get<DeviceListResponse>('/v1/devices', { params: { page, pageSize } }).then(unwrap),

  get: (id: string): Promise<Device> =>
    apiClient.get<Device>(`/v1/devices/${id}`).then(unwrap),

  action: (id: string, body: DeviceActionRequest): Promise<Device> =>
    apiClient.post<Device>(`/v1/devices/${id}/action`, body).then(unwrap),

  delete: (id: string): Promise<void> =>
    apiClient.delete(`/v1/devices/${id}`).then(() => undefined),
};

export const usersApi = {
  list: (page = 1, pageSize = 50): Promise<UserListResponse> =>
    apiClient.get<UserListResponse>('/v1/users', { params: { page, pageSize } }).then(unwrap),

  get: (id: string): Promise<AdminUser> =>
    apiClient.get<AdminUser>(`/v1/users/${id}`).then(unwrap),

  setStatus: (id: string, status: UserStatus): Promise<AdminUser> =>
    apiClient.patch<AdminUser>(`/v1/users/${id}`, { status }).then(unwrap),

  create: (username: string, email: string, displayName: string, password: string): Promise<AdminUser> =>
    apiClient.post<AdminUser>('/v1/admin/users', { username, email, displayName, password }).then(unwrap),
};

export const auditApi = {
  list: (page = 1, pageSize = 100): Promise<AuditListResponse> =>
    apiClient.get<AuditListResponse>('/v1/audit', { params: { page, pageSize } }).then(unwrap),
};

export const dashboardApi = {
  stats: (): Promise<DashboardStats> =>
    apiClient.get<DashboardStats>('/v1/dashboard/stats').then(unwrap),
};

export const policiesApi = {
  list: (): Promise<PolicyListResponse> =>
    apiClient.get<PolicyListResponse>('/v1/policies').then(unwrap),

  get: (id: string): Promise<Policy> =>
    apiClient.get<Policy>(`/v1/policies/${id}`).then(unwrap),

  create: (body: Omit<Policy, 'id' | 'createdAt' | 'updatedAt'>): Promise<Policy> =>
    apiClient.post<Policy>('/v1/policies', body).then(unwrap),

  update: (id: string, body: Partial<Omit<Policy, 'id' | 'createdAt' | 'updatedAt'>>): Promise<Policy> =>
    apiClient.put<Policy>(`/v1/policies/${id}`, body).then(unwrap),

  delete: (id: string): Promise<void> =>
    apiClient.delete(`/v1/policies/${id}`).then(() => undefined),
};

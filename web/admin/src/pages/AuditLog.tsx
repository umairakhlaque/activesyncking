import React from 'react';
import { AlertCircle, CheckCircle, Info, AlertTriangle } from 'lucide-react';

interface AuditEvent {
  id: string;
  time: string;
  action: string;
  username: string;
  deviceId: string;
  clientIp: string;
  result: string;
  severity: 'info' | 'warn' | 'error' | 'critical';
}

const events: AuditEvent[] = [
  { id: '1', time: '2024-05-06 14:32:01', action: 'AUTH_SUCCESS', username: 'jsmith', deviceId: 'iPhone14Pro-ABC123', clientIp: '10.0.1.45', result: 'TRUSTED_DEVICE', severity: 'info' },
  { id: '2', time: '2024-05-06 14:31:44', action: 'MFA_SUCCESS', username: 'alee', deviceId: 'SamsungS24-DEF456', clientIp: '10.0.1.67', result: 'MFA_VALIDATED', severity: 'info' },
  { id: '3', time: '2024-05-06 14:31:12', action: 'MFA_REQUIRED', username: 'alee', deviceId: 'SamsungS24-DEF456', clientIp: '10.0.1.67', result: 'MFA_CHALLENGE_ISSUED', severity: 'info' },
  { id: '4', time: '2024-05-06 14:28:55', action: 'MFA_FAILURE', username: 'bwilson', deviceId: 'Pixel8-GHI789', clientIp: '10.0.2.12', result: 'INVALID_OTP', severity: 'warn' },
  { id: '5', time: '2024-05-06 14:25:03', action: 'AUTH_FAILURE', username: 'unknown', deviceId: 'N/A', clientIp: '185.220.101.45', result: 'INVALID_CREDENTIALS', severity: 'warn' },
  { id: '6', time: '2024-05-06 14:10:22', action: 'DEVICE_BLOCKED', username: 'bwilson', deviceId: 'Pixel8-GHI789', clientIp: '10.0.2.12', result: 'ADMIN_ACTION', severity: 'error' },
  { id: '7', time: '2024-05-06 13:58:11', action: 'MFA_LOCKED', username: 'guest01', deviceId: 'OutlookMob-JKL012', clientIp: '203.0.113.7', result: 'MAX_ATTEMPTS_EXCEEDED', severity: 'critical' },
];

const severityIcon: Record<string, React.ReactNode> = {
  info:     <Info size={14} color="#2563eb" />,
  warn:     <AlertTriangle size={14} color="#d97706" />,
  error:    <AlertCircle size={14} color="#dc2626" />,
  critical: <AlertCircle size={14} color="#7c3aed" />,
};

const severityStyle: Record<string, React.CSSProperties> = {
  info:     { color: '#1d4ed8' },
  warn:     { color: '#b45309' },
  error:    { color: '#b91c1c' },
  critical: { color: '#6d28d9', fontWeight: 600 },
};

export default function AuditLog() {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>Audit Log</h1>
          <p style={{ margin: '4px 0 0', color: '#6b7280', fontSize: 14 }}>Complete record of authentication and access decisions</p>
        </div>
        <button style={{ background: '#fff', color: '#374151', border: '1px solid #d1d5db', padding: '8px 16px', borderRadius: 6, cursor: 'pointer', fontSize: 14 }}>
          Export CSV
        </button>
      </div>

      <div style={{ background: '#fff', border: '1px solid #e5e7eb', borderRadius: 8, overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
          <thead>
            <tr style={{ background: '#f9fafb', borderBottom: '1px solid #e5e7eb' }}>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>Time</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>Action</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>User</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>Device</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>IP</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>Result</th>
              <th style={{ padding: '10px 16px', textAlign: 'left', fontWeight: 600, color: '#374151' }}>Severity</th>
            </tr>
          </thead>
          <tbody>
            {events.map((e, i) => (
              <tr key={e.id} style={{ borderBottom: i < events.length - 1 ? '1px solid #f3f4f6' : 'none' }}>
                <td style={{ padding: '10px 16px', color: '#6b7280', fontFamily: 'monospace', fontSize: 12 }}>{e.time}</td>
                <td style={{ padding: '10px 16px' }}>
                  <code style={{ background: '#f3f4f6', padding: '2px 6px', borderRadius: 4, fontSize: 11 }}>{e.action}</code>
                </td>
                <td style={{ padding: '10px 16px', fontWeight: 500 }}>{e.username}</td>
                <td style={{ padding: '10px 16px', color: '#6b7280', fontSize: 12, fontFamily: 'monospace' }}>{e.deviceId}</td>
                <td style={{ padding: '10px 16px', color: '#6b7280', fontFamily: 'monospace', fontSize: 12 }}>{e.clientIp}</td>
                <td style={{ padding: '10px 16px', fontSize: 12 }}>{e.result}</td>
                <td style={{ padding: '10px 16px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4, ...severityStyle[e.severity] }}>
                    {severityIcon[e.severity]}
                    {e.severity}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

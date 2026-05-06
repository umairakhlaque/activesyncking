import { useState } from 'react';
import type { AuditResult, AuditSeverity } from '../api/client';

interface AuditRow {
  id: string;
  time: string;
  action: string;
  username: string;
  device: string;
  ip: string;
  result: AuditResult;
  severity: AuditSeverity;
}

const PLACEHOLDER_AUDIT: AuditRow[] = [
  { id: '1', time: '2026-05-06 09:47:12', action: 'ACTIVESYNC_AUTH', username: 'jsmith', device: 'DEV-A1B2C3', ip: '10.0.1.55', result: 'allow', severity: 'info' },
  { id: '2', time: '2026-05-06 09:45:03', action: 'DEVICE_ENROL', username: 'lchen', device: 'DEV-D4E5F6', ip: '10.0.2.12', result: 'allow', severity: 'info' },
  { id: '3', time: '2026-05-06 09:41:58', action: 'MFA_CHALLENGE_FAIL', username: 'rgarcia', device: 'DEV-G7H8I9', ip: '10.0.3.77', result: 'deny', severity: 'warning' },
  { id: '4', time: '2026-05-06 09:38:22', action: 'ACTIVESYNC_AUTH', username: 'bwilson', device: 'DEV-J0K1L2', ip: '192.168.10.8', result: 'allow', severity: 'info' },
  { id: '5', time: '2026-05-06 09:30:01', action: 'POLICY_BLOCK', username: 'unknown', device: 'DEV-M3N4O5', ip: '185.220.101.4', result: 'deny', severity: 'critical' },
  { id: '6', time: '2026-05-06 09:22:44', action: 'ACTIVESYNC_AUTH', username: 'cdavis', device: 'DEV-V2W3X4', ip: '10.0.1.200', result: 'allow', severity: 'info' },
  { id: '7', time: '2026-05-06 09:15:30', action: 'DEVICE_BLOCKED', username: 'afoster', device: 'DEV-S9T0U1', ip: '10.0.5.9', result: 'deny', severity: 'warning' },
  { id: '8', time: '2026-05-06 09:01:17', action: 'MFA_CHALLENGE_FAIL', username: 'tpatel', device: 'DEV-P6Q7R8', ip: '203.0.113.50', result: 'deny', severity: 'critical' },
  { id: '9', time: '2026-05-06 08:55:02', action: 'ACTIVESYNC_AUTH', username: 'cdavis', device: 'DEV-V2W3X4', ip: '10.0.1.200', result: 'allow', severity: 'info' },
  { id: '10', time: '2026-05-05 22:14:00', action: 'QUARANTINE_AUTO', username: 'tpatel', device: 'DEV-M3N4O5', ip: '203.0.113.50', result: 'deny', severity: 'critical' },
];

const RESULT_STYLE: Record<AuditResult, React.CSSProperties> = {
  allow: { backgroundColor: '#dcfce7', color: '#166534' },
  deny: { backgroundColor: '#fee2e2', color: '#991b1b' },
  error: { backgroundColor: '#fef3c7', color: '#92400e' },
};

const SEVERITY_STYLE: Record<AuditSeverity, React.CSSProperties> = {
  info: { backgroundColor: '#eff6ff', color: '#1d4ed8' },
  warning: { backgroundColor: '#fef9c3', color: '#854d0e' },
  critical: { backgroundColor: '#fee2e2', color: '#991b1b' },
};

const s = {
  header: { marginBottom: '20px' } as React.CSSProperties,
  pageTitle: { fontSize: '22px', fontWeight: 700, color: '#0f172a' } as React.CSSProperties,
  pageSub: { fontSize: '14px', color: '#64748b', marginTop: '4px' } as React.CSSProperties,

  filterRow: {
    display: 'flex',
    gap: '8px',
    marginBottom: '16px',
    alignItems: 'center',
  } as React.CSSProperties,

  filterBtn: (active: boolean): React.CSSProperties => ({
    padding: '5px 12px',
    fontSize: '12px',
    fontWeight: active ? 600 : 400,
    border: '1px solid',
    borderColor: active ? '#3b82f6' : '#e2e8f0',
    borderRadius: '20px',
    cursor: 'pointer',
    backgroundColor: active ? '#eff6ff' : '#fff',
    color: active ? '#1d4ed8' : '#64748b',
  }),

  tableCard: {
    backgroundColor: '#fff',
    borderRadius: '10px',
    border: '1px solid #e2e8f0',
    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
    overflow: 'hidden',
  } as React.CSSProperties,

  table: { width: '100%', borderCollapse: 'collapse' as const, fontSize: '13px' } as React.CSSProperties,

  th: {
    padding: '10px 16px',
    textAlign: 'left' as const,
    fontSize: '11px',
    fontWeight: 600,
    textTransform: 'uppercase' as const,
    letterSpacing: '0.06em',
    color: '#64748b',
    backgroundColor: '#f8fafc',
    borderBottom: '1px solid #e2e8f0',
  } as React.CSSProperties,

  td: { padding: '10px 16px', borderBottom: '1px solid #f1f5f9', color: '#334155' } as React.CSSProperties,

  pill: (style: React.CSSProperties): React.CSSProperties => ({
    display: 'inline-block',
    padding: '2px 9px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    textTransform: 'uppercase' as const,
    ...style,
  }),
};

type SeverityFilter = 'all' | AuditSeverity;

export default function AuditLog() {
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all');

  const filtered =
    severityFilter === 'all'
      ? PLACEHOLDER_AUDIT
      : PLACEHOLDER_AUDIT.filter((e) => e.severity === severityFilter);

  const severityFilters: Array<{ value: SeverityFilter; label: string }> = [
    { value: 'all', label: 'All' },
    { value: 'info', label: 'Info' },
    { value: 'warning', label: 'Warning' },
    { value: 'critical', label: 'Critical' },
  ];

  return (
    <div>
      <div style={s.header}>
        <div style={s.pageTitle}>Audit Log</div>
        <div style={s.pageSub}>Authentication and policy enforcement events</div>
      </div>

      <div style={s.filterRow}>
        <span style={{ fontSize: '12px', color: '#64748b', marginRight: '4px' }}>Severity:</span>
        {severityFilters.map(({ value, label }) => (
          <button
            key={value}
            style={s.filterBtn(severityFilter === value)}
            onClick={() => setSeverityFilter(value)}
          >
            {label}
          </button>
        ))}
      </div>

      <div style={s.tableCard}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Time</th>
              <th style={s.th}>Action</th>
              <th style={s.th}>User</th>
              <th style={s.th}>Device</th>
              <th style={s.th}>IP Address</th>
              <th style={s.th}>Result</th>
              <th style={s.th}>Severity</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((ev) => (
              <tr key={ev.id}>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px', whiteSpace: 'nowrap', color: '#64748b' }}>
                  {ev.time}
                </td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.action}</td>
                <td style={{ ...s.td, fontWeight: 500 }}>{ev.username}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.device}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.ip}</td>
                <td style={s.td}>
                  <span style={s.pill(RESULT_STYLE[ev.result])}>{ev.result}</span>
                </td>
                <td style={s.td}>
                  <span style={s.pill(SEVERITY_STYLE[ev.severity])}>{ev.severity}</span>
                </td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={7} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  No events match the current filter.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

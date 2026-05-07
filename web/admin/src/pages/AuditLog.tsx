import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { auditApi, type AuditResult, type AuditSeverity } from '../api/client';

const RESULT_STYLE: Record<AuditResult, React.CSSProperties> = {
  allow: { backgroundColor: '#dcfce7', color: '#166534' },
  deny:  { backgroundColor: '#fee2e2', color: '#991b1b' },
  error: { backgroundColor: '#fef3c7', color: '#92400e' },
};

const SEVERITY_STYLE: Record<AuditSeverity, React.CSSProperties> = {
  info:     { backgroundColor: '#eff6ff', color: '#1d4ed8' },
  warning:  { backgroundColor: '#fef9c3', color: '#854d0e' },
  critical: { backgroundColor: '#fee2e2', color: '#991b1b' },
};

const s = {
  header: { marginBottom: '20px' } as React.CSSProperties,
  pageTitle: { fontSize: '22px', fontWeight: 700, color: '#0f172a' } as React.CSSProperties,
  pageSub:   { fontSize: '14px', color: '#64748b', marginTop: '4px' } as React.CSSProperties,

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

const SEVERITY_FILTERS: Array<{ value: SeverityFilter; label: string }> = [
  { value: 'all',      label: 'All' },
  { value: 'info',     label: 'Info' },
  { value: 'warning',  label: 'Warning' },
  { value: 'critical', label: 'Critical' },
];

export default function AuditLog() {
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all');

  const { data, isLoading } = useQuery({
    queryKey: ['audit'],
    queryFn: () => auditApi.list(1, 200),
    refetchInterval: 15_000,
  });

  const allEvents = data?.items ?? [];
  const filtered = severityFilter === 'all'
    ? allEvents
    : allEvents.filter((e) => e.severity === severityFilter);

  return (
    <div>
      <div style={s.header}>
        <div style={s.pageTitle}>Audit Log</div>
        <div style={s.pageSub}>
          {isLoading ? 'Loading…' : `${data?.total ?? 0} total events`}
        </div>
      </div>

      <div style={s.filterRow}>
        <span style={{ fontSize: '12px', color: '#64748b', marginRight: '4px' }}>Severity:</span>
        {SEVERITY_FILTERS.map(({ value, label }) => (
          <button key={value} style={s.filterBtn(severityFilter === value)} onClick={() => setSeverityFilter(value)}>
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
            {isLoading && (
              <tr>
                <td colSpan={7} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  Loading…
                </td>
              </tr>
            )}
            {!isLoading && filtered.length === 0 && (
              <tr>
                <td colSpan={7} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  No events match the current filter.
                </td>
              </tr>
            )}
            {filtered.map((ev) => (
              <tr key={ev.id}>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px', whiteSpace: 'nowrap', color: '#64748b' }}>
                  {ev.timestamp}
                </td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.action}</td>
                <td style={{ ...s.td, fontWeight: 500 }}>{ev.username || '—'}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.deviceId || '—'}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.ipAddress || '—'}</td>
                <td style={s.td}>
                  <span style={s.pill(RESULT_STYLE[ev.result as AuditResult] ?? RESULT_STYLE.error)}>
                    {ev.result}
                  </span>
                </td>
                <td style={s.td}>
                  <span style={s.pill(SEVERITY_STYLE[ev.severity as AuditSeverity] ?? SEVERITY_STYLE.warning)}>
                    {ev.severity}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

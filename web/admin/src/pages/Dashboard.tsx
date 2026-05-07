import { useQuery } from '@tanstack/react-query';
import { Users, Smartphone, Clock, Activity } from 'lucide-react';
import { dashboardApi, auditApi, type AuditEvent } from '../api/client';

const s = {
  pageTitle: {
    fontSize: '22px',
    fontWeight: 700,
    color: '#0f172a',
    marginBottom: '6px',
  } as React.CSSProperties,

  pageSubtitle: {
    fontSize: '14px',
    color: '#64748b',
    marginBottom: '28px',
  } as React.CSSProperties,

  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
    gap: '16px',
    marginBottom: '32px',
  } as React.CSSProperties,

  card: {
    backgroundColor: '#fff',
    borderRadius: '10px',
    padding: '20px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
    border: '1px solid #e2e8f0',
  } as React.CSSProperties,

  cardHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: '12px',
  } as React.CSSProperties,

  cardLabel: {
    fontSize: '13px',
    color: '#64748b',
    fontWeight: 500,
  } as React.CSSProperties,

  cardIconWrap: (color: string): React.CSSProperties => ({
    width: '36px',
    height: '36px',
    borderRadius: '8px',
    backgroundColor: `${color}18`,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  }),

  cardValue: {
    fontSize: '28px',
    fontWeight: 700,
    color: '#0f172a',
    lineHeight: 1,
    marginBottom: '6px',
  } as React.CSSProperties,

  cardSub: {
    fontSize: '12px',
    color: '#94a3b8',
  } as React.CSSProperties,

  sectionTitle: {
    fontSize: '15px',
    fontWeight: 600,
    color: '#0f172a',
    marginBottom: '12px',
  } as React.CSSProperties,

  tableCard: {
    backgroundColor: '#fff',
    borderRadius: '10px',
    border: '1px solid #e2e8f0',
    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
    overflow: 'hidden',
  } as React.CSSProperties,

  table: {
    width: '100%',
    borderCollapse: 'collapse' as const,
    fontSize: '13px',
  } as React.CSSProperties,

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

  td: {
    padding: '10px 16px',
    borderBottom: '1px solid #f1f5f9',
    color: '#334155',
  } as React.CSSProperties,

  badge: (result: string): React.CSSProperties => ({
    display: 'inline-block',
    padding: '2px 8px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    backgroundColor: result === 'allow' ? '#dcfce7' : '#fee2e2',
    color: result === 'allow' ? '#166534' : '#991b1b',
  }),

  skeleton: {
    backgroundColor: '#e2e8f0',
    borderRadius: '4px',
    height: '28px',
    width: '80px',
    display: 'inline-block',
  } as React.CSSProperties,
};

const STAT_DEFS = [
  { key: 'totalUsers' as const,        label: 'Total Users',        color: '#3b82f6', Icon: Users },
  { key: 'totalDevices' as const,      label: 'Total Devices',      color: '#10b981', Icon: Smartphone },
  { key: 'pendingApprovals' as const,  label: 'Pending Approvals',  color: '#f59e0b', Icon: Clock },
  { key: 'authEventsLast24h' as const, label: 'Auth Events (24h)',  color: '#8b5cf6', Icon: Activity },
];

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: dashboardApi.stats,
    refetchInterval: 60_000,
  });

  const { data: audit } = useQuery({
    queryKey: ['audit-recent'],
    queryFn: () => auditApi.list(1, 10),
    refetchInterval: 30_000,
  });

  const recentEvents: AuditEvent[] = audit?.items ?? [];

  return (
    <div>
      <div style={s.pageTitle}>Dashboard</div>
      <div style={s.pageSubtitle}>
        Overview of your SyncGuard MFA gateway —{' '}
        {new Date().toLocaleDateString('en-GB', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
      </div>

      <div style={s.grid}>
        {STAT_DEFS.map(({ key, label, color, Icon }) => (
          <div key={key} style={s.card}>
            <div style={s.cardHeader}>
              <span style={s.cardLabel}>{label}</span>
              <div style={s.cardIconWrap(color)}>
                <Icon size={16} color={color} strokeWidth={2} />
              </div>
            </div>
            <div style={s.cardValue}>
              {statsLoading ? <span style={s.skeleton} /> : (stats?.[key] ?? 0).toLocaleString()}
            </div>
            {key === 'pendingApprovals' && !statsLoading && (
              <div style={s.cardSub}>{stats?.blockedDevices ?? 0} blocked or quarantined</div>
            )}
            {key === 'authEventsLast24h' && !statsLoading && (
              <div style={s.cardSub}>{stats?.mfaFailuresLast24h ?? 0} MFA failures</div>
            )}
          </div>
        ))}
      </div>

      <div style={s.sectionTitle}>Recent Events</div>
      <div style={s.tableCard}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Time</th>
              <th style={s.th}>Action</th>
              <th style={s.th}>User</th>
              <th style={s.th}>Device</th>
              <th style={s.th}>IP</th>
              <th style={s.th}>Result</th>
            </tr>
          </thead>
          <tbody>
            {recentEvents.length === 0 && (
              <tr>
                <td colSpan={6} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  No events yet
                </td>
              </tr>
            )}
            {recentEvents.map((ev) => (
              <tr key={ev.id}>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px', color: '#64748b', whiteSpace: 'nowrap' }}>
                  {ev.timestamp}
                </td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.action}</td>
                <td style={{ ...s.td, fontWeight: 500 }}>{ev.username || '—'}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.deviceId || '—'}</td>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{ev.ipAddress || '—'}</td>
                <td style={s.td}>
                  <span style={s.badge(ev.result)}>{ev.result.toUpperCase()}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

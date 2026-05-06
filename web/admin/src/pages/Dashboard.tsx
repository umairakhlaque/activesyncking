import { Users, Smartphone, Clock, Activity } from 'lucide-react';

interface StatCard {
  label: string;
  value: number | string;
  sub: string;
  color: string;
  Icon: React.ElementType;
}

const STATS: StatCard[] = [
  {
    label: 'Total Users',
    value: 1_284,
    sub: '+12 this week',
    color: '#3b82f6',
    Icon: Users,
  },
  {
    label: 'Total Devices',
    value: 3_472,
    sub: '98 pending approval',
    color: '#10b981',
    Icon: Smartphone,
  },
  {
    label: 'Pending Approvals',
    value: 98,
    sub: '14 flagged as suspicious',
    color: '#f59e0b',
    Icon: Clock,
  },
  {
    label: 'Auth Events (24h)',
    value: '18,430',
    sub: '34 denied, 6 MFA failures',
    color: '#8b5cf6',
    Icon: Activity,
  },
];

interface RecentEvent {
  id: string;
  time: string;
  action: string;
  user: string;
  device: string;
  result: 'allow' | 'deny';
}

const RECENT_EVENTS: RecentEvent[] = [
  { id: '1', time: '09:47:12', action: 'ActiveSync Auth', user: 'jsmith', device: 'iPhone 15 Pro', result: 'allow' },
  { id: '2', time: '09:45:03', action: 'Device Enrolment', user: 'lchen', device: 'Galaxy S24', result: 'allow' },
  { id: '3', time: '09:41:58', action: 'MFA Challenge', user: 'rgarcia', device: 'Outlook/Win11', result: 'deny' },
  { id: '4', time: '09:38:22', action: 'ActiveSync Auth', user: 'bwilson', device: 'iPad Air', result: 'allow' },
  { id: '5', time: '09:30:01', action: 'Policy Block', user: 'unknown', device: 'Unknown Android', result: 'deny' },
];

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

  badge: (result: 'allow' | 'deny'): React.CSSProperties => ({
    display: 'inline-block',
    padding: '2px 8px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    backgroundColor: result === 'allow' ? '#dcfce7' : '#fee2e2',
    color: result === 'allow' ? '#166534' : '#991b1b',
  }),
};

export default function Dashboard() {
  return (
    <div>
      <div style={s.pageTitle}>Dashboard</div>
      <div style={s.pageSubtitle}>
        Overview of your SyncGuard MFA gateway — {new Date().toLocaleDateString('en-GB', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
      </div>

      {/* Stat cards */}
      <div style={s.grid}>
        {STATS.map((stat) => (
          <div key={stat.label} style={s.card}>
            <div style={s.cardHeader}>
              <span style={s.cardLabel}>{stat.label}</span>
              <div style={s.cardIconWrap(stat.color)}>
                <stat.Icon size={16} color={stat.color} strokeWidth={2} />
              </div>
            </div>
            <div style={s.cardValue}>{typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}</div>
            <div style={s.cardSub}>{stat.sub}</div>
          </div>
        ))}
      </div>

      {/* Recent events */}
      <div style={s.sectionTitle}>Recent Events</div>
      <div style={s.tableCard}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Time</th>
              <th style={s.th}>Action</th>
              <th style={s.th}>User</th>
              <th style={s.th}>Device</th>
              <th style={s.th}>Result</th>
            </tr>
          </thead>
          <tbody>
            {RECENT_EVENTS.map((ev) => (
              <tr key={ev.id}>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px', color: '#64748b' }}>{ev.time}</td>
                <td style={s.td}>{ev.action}</td>
                <td style={{ ...s.td, fontWeight: 500 }}>{ev.user}</td>
                <td style={s.td}>{ev.device}</td>
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

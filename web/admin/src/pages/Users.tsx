import type { UserStatus, UserSource } from '../api/client';

interface UserRow {
  id: string;
  username: string;
  email: string;
  status: UserStatus;
  source: UserSource;
  mfaEnrolled: boolean;
  deviceCount: number;
}

const PLACEHOLDER_USERS: UserRow[] = [
  { id: '1', username: 'jsmith', email: 'jsmith@corp.example.com', status: 'active', source: 'ldap', mfaEnrolled: true, deviceCount: 2 },
  { id: '2', username: 'lchen', email: 'lchen@corp.example.com', status: 'active', source: 'ldap', mfaEnrolled: true, deviceCount: 1 },
  { id: '3', username: 'rgarcia', email: 'rgarcia@corp.example.com', status: 'active', source: 'entra', mfaEnrolled: false, deviceCount: 1 },
  { id: '4', username: 'bwilson', email: 'bwilson@corp.example.com', status: 'active', source: 'ldap', mfaEnrolled: true, deviceCount: 3 },
  { id: '5', username: 'tpatel', email: 'tpatel@corp.example.com', status: 'locked', source: 'ldap', mfaEnrolled: true, deviceCount: 1 },
  { id: '6', username: 'mnovak', email: 'mnovak@corp.example.com', status: 'active', source: 'local', mfaEnrolled: false, deviceCount: 1 },
  { id: '7', username: 'afoster', email: 'afoster@corp.example.com', status: 'disabled', source: 'ldap', mfaEnrolled: false, deviceCount: 1 },
  { id: '8', username: 'cdavis', email: 'cdavis@corp.example.com', status: 'active', source: 'entra', mfaEnrolled: true, deviceCount: 2 },
  { id: '9', username: 'eklein', email: 'eklein@corp.example.com', status: 'active', source: 'ldap', mfaEnrolled: true, deviceCount: 1 },
  { id: '10', username: 'pwong', email: 'pwong@corp.example.com', status: 'active', source: 'ldap', mfaEnrolled: false, deviceCount: 0 },
];

const USER_STATUS_STYLE: Record<UserStatus, React.CSSProperties> = {
  active: { backgroundColor: '#dcfce7', color: '#166534' },
  disabled: { backgroundColor: '#f1f5f9', color: '#475569' },
  locked: { backgroundColor: '#fee2e2', color: '#991b1b' },
};

const SOURCE_STYLE: Record<UserSource, React.CSSProperties> = {
  ldap: { backgroundColor: '#eff6ff', color: '#1d4ed8' },
  entra: { backgroundColor: '#faf5ff', color: '#6d28d9' },
  local: { backgroundColor: '#f0fdf4', color: '#15803d' },
};

const s = {
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-end',
    marginBottom: '20px',
  } as React.CSSProperties,

  pageTitle: { fontSize: '22px', fontWeight: 700, color: '#0f172a' } as React.CSSProperties,
  pageSub: { fontSize: '14px', color: '#64748b', marginTop: '4px' } as React.CSSProperties,

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
    textTransform: 'capitalize',
    ...style,
  }),

  mfaBadge: (enrolled: boolean): React.CSSProperties => ({
    display: 'inline-block',
    padding: '2px 9px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    backgroundColor: enrolled ? '#dcfce7' : '#fef3c7',
    color: enrolled ? '#166534' : '#92400e',
  }),
};

export default function Users() {
  return (
    <div>
      <div style={s.header}>
        <div>
          <div style={s.pageTitle}>Users</div>
          <div style={s.pageSub}>{PLACEHOLDER_USERS.length} users synced from directory</div>
        </div>
      </div>

      <div style={s.tableCard}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Username</th>
              <th style={s.th}>Email</th>
              <th style={s.th}>Status</th>
              <th style={s.th}>Source</th>
              <th style={s.th}>MFA Enrolled</th>
              <th style={s.th}>Devices</th>
            </tr>
          </thead>
          <tbody>
            {PLACEHOLDER_USERS.map((user) => (
              <tr key={user.id}>
                <td style={{ ...s.td, fontWeight: 600 }}>{user.username}</td>
                <td style={{ ...s.td, color: '#64748b' }}>{user.email}</td>
                <td style={s.td}>
                  <span style={s.pill(USER_STATUS_STYLE[user.status])}>{user.status}</span>
                </td>
                <td style={s.td}>
                  <span style={s.pill(SOURCE_STYLE[user.source])}>{user.source.toUpperCase()}</span>
                </td>
                <td style={s.td}>
                  <span style={s.mfaBadge(user.mfaEnrolled)}>
                    {user.mfaEnrolled ? 'Enrolled' : 'Not enrolled'}
                  </span>
                </td>
                <td style={{ ...s.td, textAlign: 'center' }}>{user.deviceCount}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

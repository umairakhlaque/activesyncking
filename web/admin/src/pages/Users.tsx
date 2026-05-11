import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { usersApi, type AdminUser, type UserStatus, type UserSource } from '../api/client';

const USER_STATUS_STYLE: Record<UserStatus, React.CSSProperties> = {
  active:   { backgroundColor: '#dcfce7', color: '#166534' },
  disabled: { backgroundColor: '#f1f5f9', color: '#475569' },
  locked:   { backgroundColor: '#fee2e2', color: '#991b1b' },
};

const SOURCE_STYLE: Record<UserSource, React.CSSProperties> = {
  ldap:  { backgroundColor: '#eff6ff', color: '#1d4ed8' },
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
  pageSub:   { fontSize: '14px', color: '#64748b', marginTop: '4px' } as React.CSSProperties,

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
    cursor: 'default',
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

  actionBtn: {
    padding: '3px 10px',
    fontSize: '11px',
    fontWeight: 600,
    border: '1px solid #e2e8f0',
    borderRadius: '5px',
    cursor: 'pointer',
    backgroundColor: '#f8fafc',
    color: '#475569',
    marginRight: '4px',
  } as React.CSSProperties,

  createUserCard: {
    backgroundColor: '#f8fafc',
    borderRadius: '10px',
    border: '1px solid #e2e8f0',
    padding: '16px',
    marginBottom: '20px',
  } as React.CSSProperties,

  createUserTitle: {
    fontSize: '14px',
    fontWeight: 600,
    color: '#0f172a',
    marginBottom: '12px',
  } as React.CSSProperties,

  formGroup: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr 1fr 1fr auto',
    gap: '10px',
    alignItems: 'flex-end',
  } as React.CSSProperties,

  formInput: {
    padding: '8px 10px',
    fontSize: '13px',
    border: '1px solid #e2e8f0',
    borderRadius: '5px',
    fontFamily: 'system-ui, sans-serif',
  } as React.CSSProperties,

  formLabel: {
    fontSize: '11px',
    fontWeight: 600,
    color: '#475569',
    textTransform: 'uppercase',
    marginBottom: '4px',
    display: 'block',
  } as React.CSSProperties,

  createBtn: {
    padding: '8px 14px',
    fontSize: '12px',
    fontWeight: 600,
    backgroundColor: '#10b981',
    color: '#fff',
    border: 'none',
    borderRadius: '5px',
    cursor: 'pointer',
  } as React.CSSProperties,

  message: (type: 'success' | 'error'): React.CSSProperties => ({
    padding: '8px 12px',
    borderRadius: '5px',
    fontSize: '12px',
    marginTop: '8px',
    backgroundColor: type === 'success' ? '#dcfce7' : '#fee2e2',
    color: type === 'success' ? '#166534' : '#991b1b',
  }),
};

export default function Users() {
  const qc = useQueryClient();
  const [formData, setFormData] = useState({ username: '', email: '', displayName: '', password: '' });
  const [createLoading, setCreateLoading] = useState(false);
  const [createMessage, setCreateMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => usersApi.list(1, 100),
  });

  const users: AdminUser[] = data?.items ?? [];

  const handleStatus = async (user: AdminUser, status: UserStatus) => {
    await usersApi.setStatus(user.id, status);
    qc.invalidateQueries({ queryKey: ['users'] });
  };

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateLoading(true);
    setCreateMessage(null);
    try {
      await usersApi.create(formData.username, formData.email, formData.displayName, formData.password);
      setCreateMessage({ type: 'success', text: `User "${formData.username}" created successfully` });
      setFormData({ username: '', email: '', displayName: '', password: '' });
      qc.invalidateQueries({ queryKey: ['users'] });
    } catch (err) {
      setCreateMessage({ type: 'error', text: `Failed to create user: ${err instanceof Error ? err.message : 'unknown error'}` });
    } finally {
      setCreateLoading(false);
    }
  };

  return (
    <div>
      <div style={s.header}>
        <div>
          <div style={s.pageTitle}>Users</div>
          <div style={s.pageSub}>
            {isLoading ? 'Loading…' : `${data?.total ?? 0} users`}
          </div>
        </div>
      </div>

      <div style={s.createUserCard}>
        <div style={s.createUserTitle}>Create Test User</div>
        <form onSubmit={handleCreateUser} style={s.formGroup}>
          <div>
            <label style={s.formLabel}>Username</label>
            <input
              style={s.formInput}
              type="text"
              placeholder="username"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              required
            />
          </div>
          <div>
            <label style={s.formLabel}>Email</label>
            <input
              style={s.formInput}
              type="email"
              placeholder="user@example.com"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              required
            />
          </div>
          <div>
            <label style={s.formLabel}>Display Name</label>
            <input
              style={s.formInput}
              type="text"
              placeholder="John Doe"
              value={formData.displayName}
              onChange={(e) => setFormData({ ...formData, displayName: e.target.value })}
            />
          </div>
          <div>
            <label style={s.formLabel}>Password</label>
            <input
              style={s.formInput}
              type="password"
              placeholder="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              required
            />
          </div>
          <button type="submit" style={s.createBtn} disabled={createLoading}>
            {createLoading ? 'Creating…' : 'Create'}
          </button>
        </form>
        {createMessage && <div style={s.message(createMessage.type)}>{createMessage.text}</div>}
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
              <th style={s.th}>Actions</th>
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
            {!isLoading && users.length === 0 && (
              <tr>
                <td colSpan={7} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  No users found
                </td>
              </tr>
            )}
            {users.map((user) => (
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
                <td style={s.td}>
                  {user.status !== 'active' && (
                    <button style={s.actionBtn} onClick={() => handleStatus(user, 'active')}>Enable</button>
                  )}
                  {user.status === 'active' && (
                    <button style={s.actionBtn} onClick={() => handleStatus(user, 'disabled')}>Disable</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

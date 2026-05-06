import { useState } from 'react';
import type { DeviceStatus } from '../api/client';

interface DeviceRow {
  id: string;
  deviceId: string;
  username: string;
  type: string;
  status: DeviceStatus;
  lastSeen: string;
}

const PLACEHOLDER_DEVICES: DeviceRow[] = [
  { id: '1', deviceId: 'DEV-A1B2C3', username: 'jsmith', type: 'iPhone 15 Pro (iOS 17)', status: 'approved', lastSeen: '2026-05-06 09:47' },
  { id: '2', deviceId: 'DEV-D4E5F6', username: 'lchen', type: 'Galaxy S24 (Android 14)', status: 'approved', lastSeen: '2026-05-06 09:45' },
  { id: '3', deviceId: 'DEV-G7H8I9', username: 'rgarcia', type: 'Outlook / Windows 11', status: 'pending', lastSeen: '2026-05-06 09:41' },
  { id: '4', deviceId: 'DEV-J0K1L2', username: 'bwilson', type: 'iPad Air (iPadOS 17)', status: 'approved', lastSeen: '2026-05-06 09:38' },
  { id: '5', deviceId: 'DEV-M3N4O5', username: 'tpatel', type: 'Unknown Android', status: 'quarantined', lastSeen: '2026-05-05 22:14' },
  { id: '6', deviceId: 'DEV-P6Q7R8', username: 'mnovak', type: 'Pixel 8 (Android 14)', status: 'pending', lastSeen: '2026-05-05 18:30' },
  { id: '7', deviceId: 'DEV-S9T0U1', username: 'afoster', type: 'iPhone 14 (iOS 16)', status: 'blocked', lastSeen: '2026-05-04 11:02' },
  { id: '8', deviceId: 'DEV-V2W3X4', username: 'cdavis', type: 'MacBook / Outlook 365', status: 'approved', lastSeen: '2026-05-06 08:55' },
];

const STATUS_STYLE: Record<DeviceStatus, React.CSSProperties> = {
  approved: { backgroundColor: '#dcfce7', color: '#166534' },
  pending: { backgroundColor: '#fef9c3', color: '#854d0e' },
  blocked: { backgroundColor: '#fee2e2', color: '#991b1b' },
  quarantined: { backgroundColor: '#fce7f3', color: '#9d174d' },
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

  statusBadge: (status: DeviceStatus): React.CSSProperties => ({
    display: 'inline-block',
    padding: '2px 9px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    textTransform: 'capitalize',
    ...STATUS_STYLE[status],
  }),

  actionBtn: (variant: 'approve' | 'block' | 'quarantine'): React.CSSProperties => {
    const map = {
      approve: { color: '#166534', borderColor: '#86efac', backgroundColor: '#f0fdf4' },
      block: { color: '#991b1b', borderColor: '#fca5a5', backgroundColor: '#fef2f2' },
      quarantine: { color: '#9d174d', borderColor: '#f9a8d4', backgroundColor: '#fdf4ff' },
    };
    return {
      padding: '3px 10px',
      fontSize: '11px',
      fontWeight: 600,
      border: '1px solid',
      borderRadius: '5px',
      cursor: 'pointer',
      marginRight: '4px',
      ...map[variant],
    };
  },

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
};

type FilterValue = 'all' | DeviceStatus;

export default function Devices() {
  const [filter, setFilter] = useState<FilterValue>('all');
  const [rows, setRows] = useState<DeviceRow[]>(PLACEHOLDER_DEVICES);

  const filtered = filter === 'all' ? rows : rows.filter((r) => r.status === filter);

  const handleAction = (id: string, action: 'approve' | 'block' | 'quarantine') => {
    const statusMap: Record<'approve' | 'block' | 'quarantine', DeviceStatus> = {
      approve: 'approved',
      block: 'blocked',
      quarantine: 'quarantined',
    };
    setRows((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: statusMap[action] } : r)),
    );
  };

  const filters: Array<{ value: FilterValue; label: string }> = [
    { value: 'all', label: 'All' },
    { value: 'approved', label: 'Approved' },
    { value: 'pending', label: 'Pending' },
    { value: 'blocked', label: 'Blocked' },
    { value: 'quarantined', label: 'Quarantined' },
  ];

  return (
    <div>
      <div style={s.header}>
        <div>
          <div style={s.pageTitle}>Devices</div>
          <div style={s.pageSub}>{rows.length} devices registered</div>
        </div>
      </div>

      <div style={s.filterRow}>
        {filters.map(({ value, label }) => (
          <button key={value} style={s.filterBtn(filter === value)} onClick={() => setFilter(value)}>
            {label}
          </button>
        ))}
      </div>

      <div style={s.tableCard}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Device ID</th>
              <th style={s.th}>User</th>
              <th style={s.th}>Type</th>
              <th style={s.th}>Status</th>
              <th style={s.th}>Last Seen</th>
              <th style={s.th}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((device) => (
              <tr key={device.id}>
                <td style={{ ...s.td, fontFamily: 'monospace', fontSize: '12px' }}>{device.deviceId}</td>
                <td style={{ ...s.td, fontWeight: 500 }}>{device.username}</td>
                <td style={s.td}>{device.type}</td>
                <td style={s.td}>
                  <span style={s.statusBadge(device.status)}>{device.status}</span>
                </td>
                <td style={{ ...s.td, color: '#64748b', fontSize: '12px' }}>{device.lastSeen}</td>
                <td style={s.td}>
                  {device.status !== 'approved' && (
                    <button style={s.actionBtn('approve')} onClick={() => handleAction(device.id, 'approve')}>
                      Approve
                    </button>
                  )}
                  {device.status !== 'blocked' && (
                    <button style={s.actionBtn('block')} onClick={() => handleAction(device.id, 'block')}>
                      Block
                    </button>
                  )}
                  {device.status !== 'quarantined' && (
                    <button style={s.actionBtn('quarantine')} onClick={() => handleAction(device.id, 'quarantine')}>
                      Quarantine
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={6} style={{ ...s.td, textAlign: 'center', color: '#94a3b8', padding: '32px' }}>
                  No devices match the current filter.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

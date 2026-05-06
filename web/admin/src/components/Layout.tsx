import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  Smartphone,
  Users,
  ScrollText,
  ShieldCheck,
  LogOut,
} from 'lucide-react';

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Dashboard', Icon: LayoutDashboard },
  { to: '/users', label: 'Users', Icon: Users },
  { to: '/devices', label: 'Devices', Icon: Smartphone },
  { to: '/audit', label: 'Audit Log', Icon: ScrollText },
  { to: '/policies', label: 'Policies', Icon: ShieldCheck },
] as const;

const styles = {
  root: {
    display: 'flex',
    minHeight: '100vh',
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
    backgroundColor: '#f4f5f7',
    color: '#1a1a2e',
  } as React.CSSProperties,

  sidebar: {
    width: '240px',
    flexShrink: 0,
    backgroundColor: '#0f172a',
    color: '#e2e8f0',
    display: 'flex',
    flexDirection: 'column' as const,
    minHeight: '100vh',
  } as React.CSSProperties,

  brand: {
    padding: '24px 20px 20px',
    borderBottom: '1px solid #1e293b',
  } as React.CSSProperties,

  brandLogo: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
  } as React.CSSProperties,

  brandIcon: {
    width: '32px',
    height: '32px',
    borderRadius: '8px',
    backgroundColor: '#3b82f6',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
  } as React.CSSProperties,

  brandName: {
    fontSize: '15px',
    fontWeight: 700,
    color: '#f8fafc',
    lineHeight: 1.2,
  } as React.CSSProperties,

  brandSub: {
    fontSize: '11px',
    color: '#64748b',
    marginTop: '2px',
    fontWeight: 400,
  } as React.CSSProperties,

  nav: {
    flex: 1,
    padding: '16px 0',
  } as React.CSSProperties,

  navLabel: {
    fontSize: '10px',
    fontWeight: 600,
    letterSpacing: '0.08em',
    textTransform: 'uppercase' as const,
    color: '#475569',
    padding: '0 20px 8px',
  } as React.CSSProperties,

  navFooter: {
    padding: '16px',
    borderTop: '1px solid #1e293b',
  } as React.CSSProperties,

  logoutBtn: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    width: '100%',
    padding: '8px 12px',
    background: 'none',
    border: 'none',
    borderRadius: '6px',
    color: '#94a3b8',
    fontSize: '14px',
    cursor: 'pointer',
  } as React.CSSProperties,

  main: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column' as const,
    overflow: 'hidden',
  } as React.CSSProperties,

  topbar: {
    height: '56px',
    backgroundColor: '#ffffff',
    borderBottom: '1px solid #e2e8f0',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
    padding: '0 24px',
    gap: '12px',
    flexShrink: 0,
  } as React.CSSProperties,

  adminBadge: {
    fontSize: '13px',
    color: '#475569',
  } as React.CSSProperties,

  avatar: {
    width: '32px',
    height: '32px',
    borderRadius: '50%',
    backgroundColor: '#3b82f6',
    color: '#fff',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '13px',
    fontWeight: 600,
  } as React.CSSProperties,

  content: {
    flex: 1,
    padding: '28px 32px',
    overflowY: 'auto' as const,
  } as React.CSSProperties,
};

const navLinkStyle = ({ isActive }: { isActive: boolean }): React.CSSProperties => ({
  display: 'flex',
  alignItems: 'center',
  gap: '10px',
  padding: '9px 20px',
  fontSize: '14px',
  fontWeight: isActive ? 600 : 400,
  color: isActive ? '#f8fafc' : '#94a3b8',
  backgroundColor: isActive ? '#1e3a5f' : 'transparent',
  textDecoration: 'none',
  borderLeft: isActive ? '3px solid #3b82f6' : '3px solid transparent',
  transition: 'background-color 0.15s, color 0.15s',
});

export default function Layout() {
  return (
    <div style={styles.root}>
      {/* Sidebar */}
      <aside style={styles.sidebar}>
        <div style={styles.brand}>
          <div style={styles.brandLogo}>
            <div style={styles.brandIcon}>
              <ShieldCheck size={18} color="#fff" strokeWidth={2.5} />
            </div>
            <div>
              <div style={styles.brandName}>SyncGuard MFA</div>
              <div style={styles.brandSub}>Admin Portal</div>
            </div>
          </div>
        </div>

        <nav style={styles.nav}>
          <div style={styles.navLabel}>Navigation</div>
          {NAV_ITEMS.map(({ to, label, Icon }) => (
            <NavLink key={to} to={to} style={navLinkStyle}>
              <Icon size={16} strokeWidth={1.8} />
              {label}
            </NavLink>
          ))}
        </nav>

        <div style={styles.navFooter}>
          <button
            style={styles.logoutBtn}
            onClick={() => {
              localStorage.removeItem('sg_admin_token');
              window.location.href = '/login';
            }}
          >
            <LogOut size={15} strokeWidth={1.8} />
            Sign out
          </button>
        </div>
      </aside>

      {/* Main area */}
      <div style={styles.main}>
        <header style={styles.topbar}>
          <span style={styles.adminBadge}>admin@syncguard</span>
          <div style={styles.avatar}>A</div>
        </header>

        <main style={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}

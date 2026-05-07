import { useState, type FormEvent } from 'react';
import { ShieldCheck } from 'lucide-react';
import { apiClient } from '../api/client';

const s = {
  page: {
    minHeight: '100vh',
    backgroundColor: '#0f172a',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
  } as React.CSSProperties,

  card: {
    backgroundColor: '#1e293b',
    borderRadius: '12px',
    border: '1px solid #334155',
    padding: '40px',
    width: '100%',
    maxWidth: '380px',
  } as React.CSSProperties,

  logo: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    marginBottom: '28px',
  } as React.CSSProperties,

  logoIcon: {
    width: '40px',
    height: '40px',
    borderRadius: '10px',
    backgroundColor: '#3b82f6',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
  } as React.CSSProperties,

  logoText: {
    fontSize: '18px',
    fontWeight: 700,
    color: '#f8fafc',
  } as React.CSSProperties,

  logoSub: {
    fontSize: '12px',
    color: '#64748b',
    marginTop: '2px',
  } as React.CSSProperties,

  title: {
    fontSize: '20px',
    fontWeight: 600,
    color: '#f1f5f9',
    marginBottom: '6px',
  } as React.CSSProperties,

  subtitle: {
    fontSize: '13px',
    color: '#64748b',
    marginBottom: '28px',
  } as React.CSSProperties,

  label: {
    display: 'block',
    fontSize: '13px',
    fontWeight: 500,
    color: '#94a3b8',
    marginBottom: '6px',
  } as React.CSSProperties,

  input: {
    width: '100%',
    padding: '10px 12px',
    backgroundColor: '#0f172a',
    border: '1px solid #334155',
    borderRadius: '6px',
    color: '#f1f5f9',
    fontSize: '14px',
    outline: 'none',
    boxSizing: 'border-box' as const,
  } as React.CSSProperties,

  btn: {
    width: '100%',
    padding: '11px',
    backgroundColor: '#3b82f6',
    color: '#fff',
    border: 'none',
    borderRadius: '6px',
    fontSize: '14px',
    fontWeight: 600,
    cursor: 'pointer',
    marginTop: '20px',
  } as React.CSSProperties,

  error: {
    marginTop: '12px',
    padding: '10px 12px',
    backgroundColor: '#450a0a',
    border: '1px solid #7f1d1d',
    borderRadius: '6px',
    color: '#fca5a5',
    fontSize: '13px',
  } as React.CSSProperties,
};

export default function Login() {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await apiClient.post<{ token: string }>('/v1/auth/login', { password });
      localStorage.setItem('sg_admin_token', res.data.token);
      window.location.href = '/dashboard';
    } catch {
      setError('Invalid admin password. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={s.page}>
      <div style={s.card}>
        <div style={s.logo}>
          <div style={s.logoIcon}>
            <ShieldCheck size={20} color="#fff" strokeWidth={2.5} />
          </div>
          <div>
            <div style={s.logoText}>SyncGuard MFA</div>
            <div style={s.logoSub}>Admin Portal</div>
          </div>
        </div>

        <div style={s.title}>Sign in</div>
        <div style={s.subtitle}>Enter your admin password to continue</div>

        <form onSubmit={handleSubmit}>
          <label style={s.label} htmlFor="password">Admin password</label>
          <input
            id="password"
            type="password"
            style={s.input}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••••••"
            autoFocus
            required
          />
          <button type="submit" style={s.btn} disabled={loading}>
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
          {error && <div style={s.error}>{error}</div>}
        </form>
      </div>
    </div>
  );
}

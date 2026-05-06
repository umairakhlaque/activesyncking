import React from 'react';
import { Settings, Clock } from 'lucide-react';

export default function Policies() {
  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>Policy Engine</h1>
        <p style={{ margin: '4px 0 0', color: '#6b7280', fontSize: 14 }}>
          Access policies, risk rules, and device trust configuration
        </p>
      </div>

      <div style={{
        background: '#fff',
        border: '1px solid #e5e7eb',
        borderRadius: 8,
        padding: 40,
        textAlign: 'center',
        color: '#6b7280',
      }}>
        <Settings size={48} color="#d1d5db" style={{ margin: '0 auto 16px' }} />
        <h2 style={{ margin: '0 0 8px', color: '#374151' }}>Policy Engine — Phase 2</h2>
        <p style={{ margin: '0 0 24px', maxWidth: 480, marginLeft: 'auto', marginRight: 'auto' }}>
          Risk-based access policies, device trust rules, and conditional MFA enforcement
          will be available in the Phase 2 release.
        </p>
        <div style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 8,
          background: '#eff6ff',
          color: '#1e40af',
          padding: '8px 16px',
          borderRadius: 20,
          fontSize: 13,
        }}>
          <Clock size={14} />
          Coming in Phase 2
        </div>

        <div style={{ marginTop: 40, display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, textAlign: 'left' }}>
          {[
            { title: 'IP Allowlist / Denylist', desc: 'Restrict access by network range or specific IP addresses.' },
            { title: 'Device Risk Scoring', desc: 'Score devices based on type, age, OS version, and behaviour.' },
            { title: 'Time-based Access', desc: 'Enforce MFA only outside business hours or specific time windows.' },
            { title: 'Per-User MFA Override', desc: 'Exempt specific users from MFA (with audit trail).' },
            { title: 'Group Policy', desc: 'Apply different policies to AD security groups.' },
            { title: 'Geofencing', desc: 'Block or challenge logins from unexpected geographic locations.' },
          ].map(p => (
            <div key={p.title} style={{ background: '#f9fafb', border: '1px solid #e5e7eb', borderRadius: 8, padding: 16 }}>
              <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 4, color: '#374151' }}>{p.title}</div>
              <div style={{ fontSize: 13, color: '#6b7280' }}>{p.desc}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

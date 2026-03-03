import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

const tabs = [
  { path: '/', label: 'Dashboard' },
  { path: '/create', label: 'Create' },
  { path: '/settings', label: 'Settings' },
];

export function BottomNav() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <nav style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      display: 'flex',
      borderTop: '1px solid var(--tg-theme-secondary-bg-color, #e0e0e0)',
      backgroundColor: 'var(--tg-theme-bg-color, #fff)',
      zIndex: 100,
    }}>
      {tabs.map(tab => {
        const active = location.pathname === tab.path;
        return (
          <button
            key={tab.path}
            onClick={() => navigate(tab.path)}
            style={{
              flex: 1,
              padding: '0.75rem 0',
              border: 'none',
              backgroundColor: 'transparent',
              color: active
                ? 'var(--tg-theme-button-color, #2481cc)'
                : 'var(--tg-theme-hint-color, #999)',
              fontWeight: active ? 600 : 400,
              fontSize: '0.8rem',
              cursor: 'pointer',
            }}
          >
            {tab.label}
          </button>
        );
      })}
    </nav>
  );
}

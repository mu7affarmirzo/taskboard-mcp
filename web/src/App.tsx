import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useTelegram } from './hooks/useTelegram';
import { useAuth } from './hooks/useAuth';
import { AppLayout } from './components/layout/AppLayout';
import { DashboardPage } from './pages/DashboardPage';
import { CreateTaskPage } from './pages/CreateTaskPage';
import { CardDetailPage } from './pages/CardDetailPage';
import { SettingsPage } from './pages/SettingsPage';
import { LoadingSpinner } from './components/common/LoadingSpinner';
import { ErrorMessage } from './components/common/ErrorMessage';

export function App() {
  const { initDataRaw, ready, expand } = useTelegram();
  const auth = useAuth(initDataRaw);

  React.useEffect(() => {
    ready();
    expand();
  }, [ready, expand]);

  if (auth.isLoading) {
    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        backgroundColor: 'var(--tg-theme-bg-color, #fff)',
        color: 'var(--tg-theme-text-color, #000)',
      }}>
        <LoadingSpinner />
        <p style={{ marginTop: '1rem', color: 'var(--tg-theme-hint-color, #999)' }}>
          Authenticating...
        </p>
      </div>
    );
  }

  if (auth.error) {
    return (
      <div style={{
        minHeight: '100vh',
        backgroundColor: 'var(--tg-theme-bg-color, #fff)',
        color: 'var(--tg-theme-text-color, #000)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}>
        <ErrorMessage message={auth.error} />
      </div>
    );
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/create" element={<CreateTaskPage />} />
          <Route path="/cards/:id" element={<CardDetailPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

import { Outlet } from 'react-router-dom';
import { BottomNav } from './BottomNav';

export function AppLayout() {
  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: 'var(--tg-theme-bg-color, #fff)',
      color: 'var(--tg-theme-text-color, #000)',
      paddingBottom: '60px',
    }}>
      <Outlet />
      <BottomNav />
    </div>
  );
}

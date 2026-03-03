import React from 'react';

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '2rem',
  },
  spinner: {
    width: 32,
    height: 32,
    border: '3px solid var(--tg-theme-secondary-bg-color, #f0f0f0)',
    borderTop: '3px solid var(--tg-theme-button-color, #2481cc)',
    borderRadius: '50%',
    animation: 'spin 0.8s linear infinite',
  },
};

export function LoadingSpinner() {
  return (
    <div style={styles.container}>
      <div style={styles.spinner} />
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}

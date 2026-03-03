interface Props {
  message: string;
  onRetry?: () => void;
}

export function ErrorMessage({ message, onRetry }: Props) {
  return (
    <div style={{
      padding: '1rem',
      margin: '1rem',
      borderRadius: 8,
      backgroundColor: 'var(--tg-theme-secondary-bg-color, #fff3f3)',
      color: 'var(--tg-theme-text-color, #cc0000)',
      textAlign: 'center',
    }}>
      <p>{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          style={{
            marginTop: '0.5rem',
            padding: '0.5rem 1rem',
            borderRadius: 6,
            border: 'none',
            backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
            color: 'var(--tg-theme-button-text-color, #fff)',
            cursor: 'pointer',
          }}
        >
          Retry
        </button>
      )}
    </div>
  );
}

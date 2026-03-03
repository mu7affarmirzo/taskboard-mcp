interface Props {
  message: string;
}

export function EmptyState({ message }: Props) {
  return (
    <div style={{
      padding: '2rem',
      textAlign: 'center',
      color: 'var(--tg-theme-hint-color, #999)',
    }}>
      <p>{message}</p>
    </div>
  );
}

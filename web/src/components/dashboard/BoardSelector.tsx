import type { Board } from '../../types/api';

interface Props {
  boards: Board[];
  selectedId: string | null;
  onChange: (id: string) => void;
}

export function BoardSelector({ boards, selectedId, onChange }: Props) {
  return (
    <select
      value={selectedId || ''}
      onChange={e => onChange(e.target.value)}
      style={{
        width: '100%',
        padding: '0.75rem',
        borderRadius: 8,
        border: '1px solid var(--tg-theme-secondary-bg-color, #ddd)',
        backgroundColor: 'var(--tg-theme-bg-color, #fff)',
        color: 'var(--tg-theme-text-color, #000)',
        fontSize: '1rem',
      }}
    >
      <option value="">Select a board</option>
      {boards.map(b => (
        <option key={b.id} value={b.id}>{b.name}</option>
      ))}
    </select>
  );
}

import { CardPreview } from './CardPreview';
import { useCardsList } from '../../hooks/useCards';
import { LoadingSpinner } from '../common/LoadingSpinner';
import type { List } from '../../types/api';

interface Props {
  list: List;
}

export function ListColumn({ list }: Props) {
  const { cards, loading } = useCardsList(list.id);

  return (
    <div style={{
      minWidth: 260,
      maxWidth: 300,
      flexShrink: 0,
      padding: '0.75rem',
      borderRadius: 8,
      backgroundColor: 'var(--tg-theme-secondary-bg-color, #f4f5f7)',
    }}>
      <h3 style={{
        margin: '0 0 0.5rem 0',
        fontSize: '0.9rem',
        fontWeight: 600,
      }}>
        {list.name} ({cards.length})
      </h3>
      {loading ? (
        <LoadingSpinner />
      ) : (
        cards.map(card => <CardPreview key={card.id} card={card} />)
      )}
    </div>
  );
}

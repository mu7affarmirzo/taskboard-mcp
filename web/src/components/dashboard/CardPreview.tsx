import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { Card } from '../../types/api';

interface Props {
  card: Card;
}

export function CardPreview({ card }: Props) {
  const navigate = useNavigate();

  return (
    <div
      onClick={() => navigate(`/cards/${card.id}`)}
      style={{
        padding: '0.75rem',
        marginBottom: '0.5rem',
        borderRadius: 6,
        backgroundColor: 'var(--tg-theme-bg-color, #fff)',
        boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        cursor: 'pointer',
      }}
    >
      <div style={{ fontWeight: 500 }}>{card.title}</div>
    </div>
  );
}

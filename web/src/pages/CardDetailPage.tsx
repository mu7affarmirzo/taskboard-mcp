import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useCardDetail, useDeleteCard, useAddComment } from '../hooks/useCards';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';

export function CardDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { card, loading, error, reload } = useCardDetail(id || null);
  const { remove, loading: deleting } = useDeleteCard();
  const { comment, loading: commenting } = useAddComment();
  const [commentText, setCommentText] = useState('');

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;
  if (!card) return <ErrorMessage message="Card not found" />;

  const handleDelete = async () => {
    if (confirm('Delete this card?')) {
      const ok = await remove(card.id);
      if (ok) navigate('/');
    }
  };

  const handleComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;
    const ok = await comment(card.id, commentText);
    if (ok) setCommentText('');
  };

  const sectionStyle: React.CSSProperties = {
    marginBottom: '1rem',
    padding: '0.75rem',
    borderRadius: 8,
    backgroundColor: 'var(--tg-theme-secondary-bg-color, #f4f5f7)',
  };

  const labelStyle: React.CSSProperties = {
    fontSize: '0.8rem',
    color: 'var(--tg-theme-hint-color, #999)',
    marginBottom: '0.25rem',
  };

  return (
    <div style={{ padding: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <button
          onClick={() => navigate(-1)}
          style={{
            background: 'none',
            border: 'none',
            color: 'var(--tg-theme-link-color, #2481cc)',
            cursor: 'pointer',
            fontSize: '1rem',
          }}
        >
          Back
        </button>
        <button
          onClick={handleDelete}
          disabled={deleting}
          style={{
            background: 'none',
            border: 'none',
            color: '#cc0000',
            cursor: 'pointer',
            fontSize: '0.9rem',
          }}
        >
          {deleting ? 'Deleting...' : 'Delete'}
        </button>
      </div>

      <h2 style={{ margin: '0 0 0.5rem 0', fontSize: '1.2rem' }}>{card.title}</h2>

      {card.url && (
        <a
          href={card.url}
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: 'var(--tg-theme-link-color, #2481cc)', fontSize: '0.85rem' }}
        >
          Open in Trello
        </a>
      )}

      {card.description && (
        <div style={{ ...sectionStyle, marginTop: '1rem' }}>
          <div style={labelStyle}>Description</div>
          <p style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{card.description}</p>
        </div>
      )}

      {card.due && (
        <div style={sectionStyle}>
          <div style={labelStyle}>Due Date</div>
          <p style={{ margin: 0 }}>{new Date(card.due).toLocaleDateString()}</p>
        </div>
      )}

      {card.labels && card.labels.length > 0 && (
        <div style={sectionStyle}>
          <div style={labelStyle}>Labels</div>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            {card.labels.map((l, i) => (
              <span key={i} style={{
                padding: '0.2rem 0.6rem',
                borderRadius: 12,
                backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
                color: '#fff',
                fontSize: '0.8rem',
              }}>{l}</span>
            ))}
          </div>
        </div>
      )}

      {card.members && card.members.length > 0 && (
        <div style={sectionStyle}>
          <div style={labelStyle}>Members</div>
          <p style={{ margin: 0 }}>{card.members.join(', ')}</p>
        </div>
      )}

      <div style={{ ...sectionStyle, marginTop: '1.5rem' }}>
        <div style={labelStyle}>Add Comment</div>
        <form onSubmit={handleComment} style={{ display: 'flex', gap: '0.5rem' }}>
          <input
            value={commentText}
            onChange={e => setCommentText(e.target.value)}
            placeholder="Write a comment..."
            style={{
              flex: 1,
              padding: '0.5rem',
              borderRadius: 6,
              border: '1px solid var(--tg-theme-secondary-bg-color, #ddd)',
              backgroundColor: 'var(--tg-theme-bg-color, #fff)',
              color: 'var(--tg-theme-text-color, #000)',
            }}
          />
          <button
            type="submit"
            disabled={commenting || !commentText.trim()}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: 6,
              border: 'none',
              backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
              color: 'var(--tg-theme-button-text-color, #fff)',
              cursor: 'pointer',
              opacity: commenting || !commentText.trim() ? 0.6 : 1,
            }}
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}

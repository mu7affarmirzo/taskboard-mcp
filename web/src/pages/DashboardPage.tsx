import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useBoards, useLists } from '../hooks/useBoards';
import { BoardSelector } from '../components/dashboard/BoardSelector';
import { ListColumn } from '../components/dashboard/ListColumn';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { EmptyState } from '../components/common/EmptyState';

export function DashboardPage() {
  const { boards, loading: boardsLoading, error, reload } = useBoards();
  const [selectedBoard, setSelectedBoard] = useState<string | null>(null);
  const { lists, loading: listsLoading } = useLists(selectedBoard);
  const navigate = useNavigate();

  if (boardsLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <div style={{ padding: '1rem' }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '1rem',
      }}>
        <h2 style={{ margin: 0, fontSize: '1.2rem' }}>Dashboard</h2>
        <button
          onClick={() => navigate('/create')}
          style={{
            width: 40,
            height: 40,
            borderRadius: '50%',
            border: 'none',
            backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
            color: 'var(--tg-theme-button-text-color, #fff)',
            fontSize: '1.5rem',
            cursor: 'pointer',
            lineHeight: 1,
          }}
        >
          +
        </button>
      </div>

      <BoardSelector
        boards={boards}
        selectedId={selectedBoard}
        onChange={setSelectedBoard}
      />

      {listsLoading && <LoadingSpinner />}

      {!listsLoading && selectedBoard && lists.length === 0 && (
        <EmptyState message="No lists found in this board" />
      )}

      {!listsLoading && lists.length > 0 && (
        <div style={{
          display: 'flex',
          gap: '0.75rem',
          overflowX: 'auto',
          paddingTop: '1rem',
          paddingBottom: '0.5rem',
        }}>
          {lists.map(list => (
            <ListColumn key={list.id} list={list} />
          ))}
        </div>
      )}
    </div>
  );
}

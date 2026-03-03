import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useBoards, useLists, useLabels, useMembers } from '../hooks/useBoards';
import { useCreateCard } from '../hooks/useCards';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';

export function CreateTaskPage() {
  const navigate = useNavigate();
  const { boards, loading: boardsLoading } = useBoards();
  const [boardId, setBoardId] = useState<string | null>(null);
  const { lists } = useLists(boardId);
  const { labels } = useLabels(boardId);
  const { members } = useMembers(boardId);
  const { create, loading: creating, error } = useCreateCard();

  const [listId, setListId] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dueDate, setDueDate] = useState('');
  const [selectedLabels, setSelectedLabels] = useState<string[]>([]);
  const [selectedMembers, setSelectedMembers] = useState<string[]>([]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title || !listId) return;

    const result = await create({
      list_id: listId,
      title,
      description: description || undefined,
      due_date: dueDate || undefined,
      label_ids: selectedLabels.length > 0 ? selectedLabels : undefined,
      member_ids: selectedMembers.length > 0 ? selectedMembers : undefined,
    });

    if (result) {
      navigate('/');
    }
  };

  const toggleLabel = (id: string) => {
    setSelectedLabels(prev =>
      prev.includes(id) ? prev.filter(l => l !== id) : [...prev, id]
    );
  };

  const toggleMember = (id: string) => {
    setSelectedMembers(prev =>
      prev.includes(id) ? prev.filter(m => m !== id) : [...prev, id]
    );
  };

  if (boardsLoading) return <LoadingSpinner />;

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '0.75rem',
    borderRadius: 8,
    border: '1px solid var(--tg-theme-secondary-bg-color, #ddd)',
    backgroundColor: 'var(--tg-theme-bg-color, #fff)',
    color: 'var(--tg-theme-text-color, #000)',
    fontSize: '1rem',
    boxSizing: 'border-box',
  };

  return (
    <div style={{ padding: '1rem' }}>
      <h2 style={{ margin: '0 0 1rem 0', fontSize: '1.2rem' }}>Create Task</h2>

      {error && <ErrorMessage message={error} />}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '0.75rem' }}>
          <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Board</label>
          <select style={inputStyle} value={boardId || ''} onChange={e => { setBoardId(e.target.value); setListId(''); }}>
            <option value="">Select board</option>
            {boards.map(b => <option key={b.id} value={b.id}>{b.name}</option>)}
          </select>
        </div>

        <div style={{ marginBottom: '0.75rem' }}>
          <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>List</label>
          <select style={inputStyle} value={listId} onChange={e => setListId(e.target.value)} disabled={!boardId}>
            <option value="">Select list</option>
            {lists.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
          </select>
        </div>

        <div style={{ marginBottom: '0.75rem' }}>
          <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Title *</label>
          <input style={inputStyle} value={title} onChange={e => setTitle(e.target.value)} placeholder="Task title" required />
        </div>

        <div style={{ marginBottom: '0.75rem' }}>
          <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Description</label>
          <textarea style={{ ...inputStyle, minHeight: 80, resize: 'vertical' }} value={description} onChange={e => setDescription(e.target.value)} placeholder="Description (optional)" />
        </div>

        <div style={{ marginBottom: '0.75rem' }}>
          <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Due Date</label>
          <input style={inputStyle} type="date" value={dueDate} onChange={e => setDueDate(e.target.value)} />
        </div>

        {labels.length > 0 && (
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Labels</label>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
              {labels.map(l => (
                <button
                  key={l.id}
                  type="button"
                  onClick={() => toggleLabel(l.id)}
                  style={{
                    padding: '0.4rem 0.8rem',
                    borderRadius: 16,
                    border: selectedLabels.includes(l.id) ? '2px solid var(--tg-theme-button-color, #2481cc)' : '1px solid var(--tg-theme-secondary-bg-color, #ddd)',
                    backgroundColor: l.color || 'var(--tg-theme-secondary-bg-color, #f0f0f0)',
                    color: '#fff',
                    fontSize: '0.85rem',
                    cursor: 'pointer',
                  }}
                >
                  {l.name || l.color}
                </button>
              ))}
            </div>
          </div>
        )}

        {members.length > 0 && (
          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Members</label>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
              {members.map(m => (
                <button
                  key={m.id}
                  type="button"
                  onClick={() => toggleMember(m.id)}
                  style={{
                    padding: '0.4rem 0.8rem',
                    borderRadius: 16,
                    border: selectedMembers.includes(m.id) ? '2px solid var(--tg-theme-button-color, #2481cc)' : '1px solid var(--tg-theme-secondary-bg-color, #ddd)',
                    backgroundColor: selectedMembers.includes(m.id) ? 'var(--tg-theme-button-color, #2481cc)' : 'transparent',
                    color: selectedMembers.includes(m.id) ? 'var(--tg-theme-button-text-color, #fff)' : 'var(--tg-theme-text-color, #000)',
                    fontSize: '0.85rem',
                    cursor: 'pointer',
                  }}
                >
                  {m.full_name || m.username}
                </button>
              ))}
            </div>
          </div>
        )}

        <button
          type="submit"
          disabled={creating || !title || !listId}
          style={{
            width: '100%',
            padding: '0.75rem',
            borderRadius: 8,
            border: 'none',
            backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
            color: 'var(--tg-theme-button-text-color, #fff)',
            fontSize: '1rem',
            fontWeight: 600,
            cursor: creating ? 'not-allowed' : 'pointer',
            opacity: creating || !title || !listId ? 0.6 : 1,
          }}
        >
          {creating ? 'Creating...' : 'Create Task'}
        </button>
      </form>
    </div>
  );
}

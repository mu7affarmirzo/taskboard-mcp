import React, { useState, useEffect } from 'react';
import { fetchSettings, updateSettings, connectTrello } from '../api/settings';
import { useBoards, useLists } from '../hooks/useBoards';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';

export function SettingsPage() {
  const [trelloConnected, setTrelloConnected] = useState(false);
  const [defaultBoardId, setDefaultBoardId] = useState('');
  const [defaultListId, setDefaultListId] = useState('');
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState('');

  const { boards } = useBoards();
  const { lists } = useLists(defaultBoardId || null);

  useEffect(() => {
    fetchSettings()
      .then(s => {
        setTrelloConnected(s.trello_connected);
        setDefaultBoardId(s.default_board_id);
        setDefaultListId(s.default_list_id);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const handleConnectTrello = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token.trim()) return;
    setSaving(true);
    setError(null);
    try {
      await connectTrello(token);
      setTrelloConnected(true);
      setToken('');
      setSuccess('Trello connected successfully');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to connect');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveDefaults = async () => {
    setSaving(true);
    setError(null);
    try {
      await updateSettings(defaultBoardId, defaultListId);
      setSuccess('Settings saved');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <LoadingSpinner />;

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

  const sectionStyle: React.CSSProperties = {
    marginBottom: '1.5rem',
    padding: '1rem',
    borderRadius: 8,
    backgroundColor: 'var(--tg-theme-secondary-bg-color, #f4f5f7)',
  };

  return (
    <div style={{ padding: '1rem' }}>
      <h2 style={{ margin: '0 0 1rem 0', fontSize: '1.2rem' }}>Settings</h2>

      {error && <ErrorMessage message={error} />}
      {success && (
        <div style={{
          padding: '0.75rem',
          marginBottom: '1rem',
          borderRadius: 8,
          backgroundColor: '#e6f9e6',
          color: '#2d7a2d',
          textAlign: 'center',
        }}>
          {success}
        </div>
      )}

      <div style={sectionStyle}>
        <h3 style={{ margin: '0 0 0.75rem 0', fontSize: '1rem' }}>Trello Connection</h3>
        {trelloConnected ? (
          <div style={{ color: 'var(--tg-theme-hint-color, #999)' }}>
            Connected
          </div>
        ) : (
          <form onSubmit={handleConnectTrello}>
            <input
              style={{ ...inputStyle, marginBottom: '0.5rem' }}
              value={token}
              onChange={e => setToken(e.target.value)}
              placeholder="Paste your Trello token"
              type="password"
            />
            <button
              type="submit"
              disabled={saving || !token.trim()}
              style={{
                width: '100%',
                padding: '0.75rem',
                borderRadius: 8,
                border: 'none',
                backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
                color: 'var(--tg-theme-button-text-color, #fff)',
                fontSize: '1rem',
                cursor: 'pointer',
                opacity: saving || !token.trim() ? 0.6 : 1,
              }}
            >
              {saving ? 'Connecting...' : 'Connect Trello'}
            </button>
          </form>
        )}
      </div>

      {trelloConnected && (
        <div style={sectionStyle}>
          <h3 style={{ margin: '0 0 0.75rem 0', fontSize: '1rem' }}>Default Board & List</h3>

          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Default Board</label>
            <select
              style={inputStyle}
              value={defaultBoardId}
              onChange={e => { setDefaultBoardId(e.target.value); setDefaultListId(''); }}
            >
              <option value="">Select board</option>
              {boards.map(b => <option key={b.id} value={b.id}>{b.name}</option>)}
            </select>
          </div>

          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--tg-theme-hint-color, #999)' }}>Default List</label>
            <select
              style={inputStyle}
              value={defaultListId}
              onChange={e => setDefaultListId(e.target.value)}
              disabled={!defaultBoardId}
            >
              <option value="">Select list</option>
              {lists.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </div>

          <button
            onClick={handleSaveDefaults}
            disabled={saving}
            style={{
              width: '100%',
              padding: '0.75rem',
              borderRadius: 8,
              border: 'none',
              backgroundColor: 'var(--tg-theme-button-color, #2481cc)',
              color: 'var(--tg-theme-button-text-color, #fff)',
              fontSize: '1rem',
              cursor: 'pointer',
              opacity: saving ? 0.6 : 1,
            }}
          >
            {saving ? 'Saving...' : 'Save Defaults'}
          </button>
        </div>
      )}
    </div>
  );
}

import { useState, useEffect, useCallback } from 'react';
import { fetchBoards, fetchLists, fetchLabels, fetchMembers } from '../api/boards';
import type { Board, List, Label, Member } from '../types/api';

export function useBoards() {
  const [boards, setBoards] = useState<Board[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetchBoards();
      setBoards(res.boards || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load boards');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return { boards, loading, error, reload: load };
}

export function useLists(boardId: string | null) {
  const [lists, setLists] = useState<List[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!boardId) { setLists([]); return; }
    setLoading(true);
    fetchLists(boardId)
      .then(res => setLists(res.lists || []))
      .catch(() => setLists([]))
      .finally(() => setLoading(false));
  }, [boardId]);

  return { lists, loading };
}

export function useLabels(boardId: string | null) {
  const [labels, setLabels] = useState<Label[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!boardId) { setLabels([]); return; }
    setLoading(true);
    fetchLabels(boardId)
      .then(res => setLabels(res.labels || []))
      .catch(() => setLabels([]))
      .finally(() => setLoading(false));
  }, [boardId]);

  return { labels, loading };
}

export function useMembers(boardId: string | null) {
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!boardId) { setMembers([]); return; }
    setLoading(true);
    fetchMembers(boardId)
      .then(res => setMembers(res.members || []))
      .catch(() => setMembers([]))
      .finally(() => setLoading(false));
  }, [boardId]);

  return { members, loading };
}

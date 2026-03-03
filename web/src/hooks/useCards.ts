import { useState, useEffect, useCallback } from 'react';
import { fetchCards } from '../api/boards';
import { fetchCard, createCard, deleteCard as apiDeleteCard, addComment as apiAddComment } from '../api/cards';
import type { Card, CardDetail, CreateCardRequest } from '../types/api';

export function useCardsList(listId: string | null) {
  const [cards, setCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    if (!listId) { setCards([]); return; }
    setLoading(true);
    try {
      const res = await fetchCards(listId);
      setCards(res.cards || []);
    } catch {
      setCards([]);
    } finally {
      setLoading(false);
    }
  }, [listId]);

  useEffect(() => { load(); }, [load]);

  return { cards, loading, reload: load };
}

export function useCardDetail(cardId: string | null) {
  const [card, setCard] = useState<CardDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!cardId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetchCard(cardId);
      setCard(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load card');
    } finally {
      setLoading(false);
    }
  }, [cardId]);

  useEffect(() => { load(); }, [load]);

  return { card, loading, error, reload: load };
}

export function useCreateCard() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const create = useCallback(async (data: CreateCardRequest) => {
    setLoading(true);
    setError(null);
    try {
      const res = await createCard(data);
      return res;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create card');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  return { create, loading, error };
}

export function useDeleteCard() {
  const [loading, setLoading] = useState(false);

  const remove = useCallback(async (cardId: string) => {
    setLoading(true);
    try {
      await apiDeleteCard(cardId);
      return true;
    } catch {
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  return { remove, loading };
}

export function useAddComment() {
  const [loading, setLoading] = useState(false);

  const comment = useCallback(async (cardId: string, text: string) => {
    setLoading(true);
    try {
      await apiAddComment(cardId, text);
      return true;
    } catch {
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  return { comment, loading };
}

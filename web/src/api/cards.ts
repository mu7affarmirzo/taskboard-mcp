import { apiFetch } from './client';
import type {
  CardDetail,
  CreateCardRequest,
  CreateCardResponse,
  UpdateCardRequest,
} from '../types/api';

export function fetchCard(cardId: string): Promise<CardDetail> {
  return apiFetch<CardDetail>(`/api/cards/${cardId}`);
}

export function createCard(data: CreateCardRequest): Promise<CreateCardResponse> {
  return apiFetch<CreateCardResponse>('/api/cards', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateCard(cardId: string, data: UpdateCardRequest): Promise<void> {
  return apiFetch(`/api/cards/${cardId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteCard(cardId: string): Promise<void> {
  return apiFetch(`/api/cards/${cardId}`, { method: 'DELETE' });
}

export function addComment(cardId: string, text: string): Promise<void> {
  return apiFetch(`/api/cards/${cardId}/comments`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  });
}

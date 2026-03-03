import { apiFetch } from './client';
import type { SettingsResponse } from '../types/api';

export function fetchSettings(): Promise<SettingsResponse> {
  return apiFetch<SettingsResponse>('/api/user/settings');
}

export function updateSettings(boardId: string, listId: string): Promise<void> {
  return apiFetch('/api/user/settings', {
    method: 'PUT',
    body: JSON.stringify({ board_id: boardId, list_id: listId }),
  });
}

export function connectTrello(token: string): Promise<void> {
  return apiFetch('/api/user/token', {
    method: 'PUT',
    body: JSON.stringify({ token }),
  });
}

import { apiFetch } from './client';
import type { AuthResponse } from '../types/api';

export function authenticate(initData: string): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/api/auth', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  });
}

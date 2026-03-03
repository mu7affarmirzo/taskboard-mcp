import { useState, useEffect, useCallback } from 'react';
import { authenticate } from '../api/auth';
import { setToken } from '../api/client';

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  userId: number | null;
  username: string | null;
}

export function useAuth(initDataRaw: string | undefined) {
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    error: null,
    userId: null,
    username: null,
  });

  const login = useCallback(async (initData: string) => {
    try {
      setState(s => ({ ...s, isLoading: true, error: null }));
      const res = await authenticate(initData);
      setToken(res.token);
      setState({
        isAuthenticated: true,
        isLoading: false,
        error: null,
        userId: res.user_id,
        username: res.username,
      });
    } catch (err) {
      setState({
        isAuthenticated: false,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Authentication failed',
        userId: null,
        username: null,
      });
    }
  }, []);

  useEffect(() => {
    if (initDataRaw) {
      login(initDataRaw);
    } else {
      setState(s => ({ ...s, isLoading: false, error: 'No init data available' }));
    }
  }, [initDataRaw, login]);

  return state;
}

import { apiFetch } from './client';
import type {
  BoardsResponse,
  ListsResponse,
  LabelsResponse,
  MembersResponse,
  CardsResponse,
} from '../types/api';

export function fetchBoards(): Promise<BoardsResponse> {
  return apiFetch<BoardsResponse>('/api/boards');
}

export function fetchLists(boardId: string): Promise<ListsResponse> {
  return apiFetch<ListsResponse>(`/api/boards/${boardId}/lists`);
}

export function fetchLabels(boardId: string): Promise<LabelsResponse> {
  return apiFetch<LabelsResponse>(`/api/boards/${boardId}/labels`);
}

export function fetchMembers(boardId: string): Promise<MembersResponse> {
  return apiFetch<MembersResponse>(`/api/boards/${boardId}/members`);
}

export function fetchCards(listId: string): Promise<CardsResponse> {
  return apiFetch<CardsResponse>(`/api/lists/${listId}/cards`);
}

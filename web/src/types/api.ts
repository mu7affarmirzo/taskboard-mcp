export interface AuthResponse {
  token: string;
  user_id: number;
  first_name: string;
  username: string;
}

export interface SettingsResponse {
  trello_connected: boolean;
  default_board_id: string;
  default_list_id: string;
}

export interface Board {
  id: string;
  name: string;
}

export interface BoardsResponse {
  boards: Board[];
}

export interface List {
  id: string;
  name: string;
}

export interface ListsResponse {
  lists: List[];
}

export interface Label {
  id: string;
  name: string;
  color: string;
}

export interface LabelsResponse {
  labels: Label[];
}

export interface Member {
  id: string;
  username: string;
  full_name: string;
}

export interface MembersResponse {
  members: Member[];
}

export interface Card {
  id: string;
  title: string;
  url: string;
  list_id: string;
}

export interface CardsResponse {
  cards: Card[];
}

export interface CardDetail {
  id: string;
  title: string;
  description: string;
  url: string;
  list_id: string;
  due: string;
  labels: string[];
  members: string[];
}

export interface CreateCardRequest {
  list_id: string;
  title: string;
  description?: string;
  due_date?: string;
  label_ids?: string[];
  member_ids?: string[];
}

export interface CreateCardResponse {
  card_id: string;
  card_url: string;
  title: string;
}

export interface UpdateCardRequest {
  title?: string;
  description?: string;
  list_id?: string;
  due?: string;
  label_ids?: string;
  member_ids?: string;
}

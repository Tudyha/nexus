export interface Workspace {
  id: number;
  name: string;
  description: string;
  status: number;
  created_at: string;
}

export interface WorkspaceUser {
  id: number;
  workspace_id: number;
  user_id: number;
  role: number;
  nickname: string;
  avatar: string;
  created_at: string;
}

export interface App {
  id: number;
  workspace_id: number;
  name: string;
  app_secret: string;
  description: string;
  status: number;
  config: string;
  created_at: string;
}

export type LoginRequest = {
  login_type: number;
  username: string;
  password?: string;
  code?: string;
};

export type LoginResponse = { token: string };

export type UserResponse = {
  id: string;
  nickname: string;
  avatar: string;
  workspace_list: WorkspaceResponse[];
};

export type WorkspaceResponse = {
  id: string;
  name: string;
  description: string;
  app_list: AppResponse[];
};

export type AppResponse = {
  id: string;
  name: string;
  description: string;
};

import http from "./index";
import type { Workspace, WorkspaceUser, App } from "@/types";

export async function getWorkspaces(): Promise<Workspace[]> {
  return http.get("/v1/workspace");
}

export async function getWorkspace(id: number): Promise<Workspace> {
  return http.get(`/v1/workspace/${id}`);
}

export async function createWorkspace(data: { name: string; description?: string }): Promise<Workspace> {
  return http.post("/v1/workspace", data);
}

export async function updateWorkspace(id: number, data: { name: string; description?: string }): Promise<void> {
  return http.put(`/v1/workspace/${id}`, data);
}

export async function deleteWorkspace(id: number): Promise<void> {
  return http.delete(`/v1/workspace/${id}`);
}

// Workspace users
export async function getWorkspaceUsers(id: number): Promise<WorkspaceUser[]> {
  return http.get(`/v1/workspace/${id}/users`);
}

export async function addWorkspaceUser(id: number, data: { user_id: number; role: number }): Promise<void> {
  return http.post(`/v1/workspace/${id}/users`, data);
}

export async function removeWorkspaceUser(workspaceId: number, userId: number): Promise<void> {
  return http.delete(`/v1/workspace/${workspaceId}/users/${userId}`);
}

// Apps
export async function getApps(workspaceId: number): Promise<App[]> {
  return http.get(`/v1/workspace/${workspaceId}/apps`);
}

export async function createApp(workspaceId: number, data: { name: string; description?: string }): Promise<App> {
  return http.post(`/v1/workspace/${workspaceId}/apps`, data);
}

export async function getApp(id: number): Promise<App> {
  return http.get(`/v1/app/${id}`);
}

export async function updateApp(id: number, data: { name: string; description?: string }): Promise<void> {
  return http.put(`/v1/app/${id}`, data);
}

export async function deleteApp(id: number): Promise<void> {
  return http.delete(`/v1/app/${id}`);
}

export async function updateAppConfig(id: number, data: { config: string }): Promise<void> {
  return http.put(`/v1/app/${id}/config`, data);
}

import http from "./index";
import type { PageResponse, VersionResponse, TaskExecutionResponse } from "@/types";

export async function getVersionPage(params: Record<string, any>): Promise<PageResponse<VersionResponse>> {
  return http.get("/v1/version/page", params);
}

export async function uploadVersion(data: FormData): Promise<void> {
  return http.post("/v1/version", data, {
    headers: { "Content-Type": "multipart/form-data" },
    timeout: 120000,
  });
}

export async function deleteVersion(id: number): Promise<void> {
  return http.delete(`/v1/version/${id}`);
}

export async function triggerUpgrade(clientId: number, force: boolean = false): Promise<{ task_id: number }> {
  return http.post(`/v1/client/${clientId}/upgrade?force=${force}`);
}

export async function getTaskProgress(taskId: number, clientId: number): Promise<TaskExecutionResponse | null> {
  return http.get(`/v1/task/${taskId}/${clientId}`);
}

import http from "./index";
import type { PageResponse, ClientResponse, ClientBindResponse, ClientSystemInfoResponse, ClientTunnelResponse } from "@/types";


export async function getClientPage(params: Record<string, any>): Promise<PageResponse<ClientResponse>> {
  return http.get("/v1/client/page", params);
}

export async function getOnlineClients(): Promise<{ id: number; hostname: string }[]> {
  return http.get("/v1/client/online");
}

export async function getClientBind(): Promise<ClientBindResponse> {
  return http.get("/v1/client/bind");
}

export async function deleteClient(id: number): Promise<void> {
  return http.delete(`/v1/client/${id}`);
}

export async function getClientDetail(id: string): Promise<ClientResponse> {
  return http.get(`/v1/client/${id}`);
}

export async function getClientSystemInfo(id: number): Promise<ClientSystemInfoResponse[]> {
  return http.get(`/v1/client/${id}/metric`);
}

export async function getV2raySubscribe(ids: number[]): Promise<string> {
  return http.post(`/v1/client/v2ray/sub`, {ids: ids});
}

export async function getClientTunnel(id: number): Promise<ClientTunnelResponse[]> {
  return http.get(`/v1/client/${id}/tunnel`);
}

export async function createClientTunnel(data: Record<string, any>): Promise<void> {
  return http.post(`/v1/tunnel`, data);
}

export async function deleteClientTunnel(clientId: number, tunnelId: number): Promise<void> {
  return http.delete(`/v1/client/${clientId}/tunnel/${tunnelId}`);
}

export async function getAllTunnels(): Promise<ClientTunnelResponse[]> {
  return http.get("/v1/tunnel/list");
}

export async function upgradeClient(id: number): Promise<void> {
  return http.post(`/v1/client/${id}/upgrade`);
}

export async function createBatchTask(taskType: number, clientIds: number[]): Promise<any[]> {
  return http.post("/v1/task/create", { task_type: taskType, client_ids: clientIds });
}

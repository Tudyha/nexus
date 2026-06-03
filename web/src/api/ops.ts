import http from "./index";
import type { FileEntry, ProcessEntry, NetworkEntry } from "@/types";

export async function getMetrics(): Promise<string> {
  return http.get("/metrics");
}

export async function getAppConfig(): Promise<{ config: string }> {
  return http.get("/v1/client/config");
}

export async function updateAppConfig(data: { config: string }): Promise<void> {
  return http.put("/v1/client/config", data);
}

const TEXT_EXTS = ['txt','md','json','xml','yaml','yml','js','ts','html','css','go','py','java','sh','log','conf','cfg','ini','env','c','cpp','h','hpp','sql','rb','php','pl','lua','rs','toml','gradle','makefile','dockerfile','gitignore','editorconfig','vue','svelte','jsx','tsx','sass','scss','less','csv','bat','cmd','ps1','zsh','bash','swift','kt','scala','erl','ex','exs','clj','groovy','gradle','cmake','makefile']
const IMG_EXTS = ['png','jpg','jpeg','gif','bmp','svg','webp','ico']

/** 获取文件类型：text / image / binary */
function detectFileType(name: string): 'text' | 'image' | 'binary' {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  if (IMG_EXTS.includes(ext)) return 'image'
  if (TEXT_EXTS.includes(ext) || ext.length === 0 || name === '') return 'text'
  return 'binary'
}

export async function fetchFile(clientId: number, filePath: string): Promise<{ type: 'text'; content: string } | { type: 'image'; blobUrl: string } | { type: 'binary' }> {
  const token = JSON.parse(localStorage.getItem("user") || "{}")?.token || ""
  const baseURL = import.meta.env.VITE_API_BASE_URL || ""
  const url = `${baseURL}/v1/client/${clientId}/files/download?path=${encodeURIComponent(filePath)}`

  const type = detectFileType(filePath)
  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } })
  if (!res.ok) throw new Error('fetch failed')

  if (type === 'image') {
    const blob = await res.blob()
    return { type: 'image', blobUrl: window.URL.createObjectURL(blob) }
  }
  if (type === 'text') {
    const content = await res.text()
    // 限制最大显示 1MB 文本
    if (content.length > 1 * 1024 * 1024) {
      return { type: 'text', content: content.slice(0, 1 * 1024 * 1024) + '\n\n... (文件过大，仅显示前 1MB)' }
    }
    return { type: 'text', content }
  }
  return { type: 'binary' }
}

// ---- 文件管理 ----

export async function getFileList(clientId: number, path: string = "/"): Promise<FileEntry[]> {
  return http.get(`/v1/client/${clientId}/files`, { path });
}

/** 直接触发浏览器下载（不经过 http service，因为需要 blob 响应） */
export function downloadFile(clientId: number, filePath: string) {
  const token = JSON.parse(localStorage.getItem("user") || "{}")?.token || ""
  const baseURL = import.meta.env.VITE_API_BASE_URL || ""
  const url = `${baseURL}/v1/client/${clientId}/files/download?path=${encodeURIComponent(filePath)}`
  const a = document.createElement("a")
  a.href = url
  a.style.display = "none"
  if (token) {
    // Use fetch with auth header to support token-based auth
    fetch(url, { headers: { Authorization: `Bearer ${token}` } })
      .then(res => res.blob())
      .then(blob => {
        const blobUrl = window.URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.href = blobUrl
        link.download = filePath.split("/").pop() || "download"
        document.body.appendChild(link)
        link.click()
        link.remove()
        window.URL.revokeObjectURL(blobUrl)
      })
      .catch(() => {})
    return
  }
  a.click()
  a.remove()
}

export async function uploadFile(clientId: number, path: string, data: Blob, append: boolean = false): Promise<void> {
  const token = JSON.parse(localStorage.getItem("user") || "{}")?.token || ""
  const baseURL = import.meta.env.VITE_API_BASE_URL || ""
  const url = `${baseURL}/v1/client/${clientId}/files/upload?path=${encodeURIComponent(path)}&append=${append}`
  const res = await fetch(url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/octet-stream",
    },
    body: data,
  })
  if (!res.ok) throw new Error("upload failed")
}

export async function createDir(clientId: number, path: string): Promise<void> {
  const token = JSON.parse(localStorage.getItem("user") || "{}")?.token || ""
  const baseURL = import.meta.env.VITE_API_BASE_URL || ""
  const url = `${baseURL}/v1/client/${clientId}/files/mkdir?path=${encodeURIComponent(path)}`
  const res = await fetch(url, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error("mkdir failed")
}

export async function renameFile(clientId: number, oldPath: string, newPath: string): Promise<void> {
  return http.post(`/v1/client/${clientId}/files/rename`, { old_path: oldPath, new_path: newPath });
}

export async function deleteFile(clientId: number, path: string): Promise<void> {
  const token = JSON.parse(localStorage.getItem("user") || "{}")?.token || ""
  const baseURL = import.meta.env.VITE_API_BASE_URL || ""
  const url = `${baseURL}/v1/client/${clientId}/files/delete?path=${encodeURIComponent(path)}`
  const res = await fetch(url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error("delete failed")
}

// ---- 进程管理 ----

export async function getProcessList(clientId: number): Promise<ProcessEntry[]> {
  return http.get(`/v1/client/${clientId}/processes`);
}

export async function killProcess(clientId: number, pid: number): Promise<void> {
  return http.delete(`/v1/client/${clientId}/processes/${pid}`);
}

// ---- 网络管理 ----

export async function getNetworkList(clientId: number): Promise<NetworkEntry[]> {
  return http.get(`/v1/client/${clientId}/network`);
}

// ---- 应用管理 ----

export interface AppTemplate {
  id: string;
  name: string;
  category: string;
  icon: string;
  description: string;
}

export interface AppEntry {
  app_id: string;
  name: string;
  version: string;
  category: string;
  icon: string;
}

export async function getAppTemplates(): Promise<AppTemplate[]> {
  return http.get("/v1/apps/templates");
}

export async function checkClientApps(clientId: number): Promise<AppEntry[]> {
  return http.get(`/v1/client/${clientId}/apps/check`);
}

export async function installClientApp(clientId: number, appId: string): Promise<void> {
  return http.post(`/v1/client/${clientId}/apps/install`, { app_id: appId });
}

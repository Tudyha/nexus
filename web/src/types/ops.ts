export interface FileEntry {
  name: string;
  size: number;
  mod_time: number;
  mode: string;
  is_dir: boolean;
  owner?: string;
  group?: string;
}

export interface ProcessEntry {
  pid: number;
  name: string;
  ppid: number;
  threads: number;
  user: string;
  cpu_percent: number;
  mem_percent: number;
  mem_rss: number;
  status: string;
  start_time: string;
}

export interface NetworkEntry {
  pid: number;
  process_name: string;
  protocol: string;
  local_addr: string;
  remote_addr: string;
  status: string;
}

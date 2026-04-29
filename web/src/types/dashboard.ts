export type DashboardResponse = {
    sys_info?: SystemInfo;
    client_stats?: ClientStats;
    sys_stats?: SysStats;
}

export type SystemInfo = {
    hostname: string;
    username: string;
    gid: string;
    uid: string;
    os_name: string;
    arch: string;
    uptime: number;
    boot_time: number;
    os: string;
    platform: string;
    platform_family: string;
    platform_version: string;
    kernel_version: string;
    kernel_arch: string;
    host_id: string;
    cpu_num: number;
    cpu_info: string;
    mem_total: number;
    disk_total: number;
}

export type ClientStats = {
    online: number;
    offline: number;
}

export type SysStats = {
    cpu_usage: number;
    mem_usage: number;
}
export const clientStatusMap = {
  1: '在线',
  2: '离线',
}

export const clientOsIconMap: Record<string, string> = {
  "linux": "mdi:linux",
  "darwin": "mdi:apple",
  "windows": "mdi:microsoft-windows",
}



export const clientOsColorMap: Record<string, string> = {
  "linux": 'bg-sky-600',
  "darwin": 'bg-indigo-500',
  "windows": 'bg-lime-500',
}

export const osOptions = [
  { value: "linux", label: "Linux" },
  { value: "darwin", label: "macOS" },
  { value: "windows", label: "Windows" },
]

export const archOptions = [
  { value: "amd64", label: "amd64" },
  { value: "arm64", label: "arm64" },
]
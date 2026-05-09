export type VersionResponse = {
  id: number;
  version: number;
  version_name: string;
  os: string;
  arch: string;
  checksum: string;
  binary_size: number;
  changelog: string;
  file_name: string;
  created_at: number;
};

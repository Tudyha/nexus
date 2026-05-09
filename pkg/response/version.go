package response

import "time"

type VersionResponse struct {
	ID          uint64    `json:"id"`
	Version     uint32    `json:"version"`
	VersionName string    `json:"version_name"`
	Os          string    `json:"os"`
	Arch        string    `json:"arch"`
	Checksum    string    `json:"checksum"`
	BinarySize  int64     `json:"binary_size"`
	Changelog   string    `json:"changelog"`
	FileName    string    `json:"file_name"`
	CreatedAt   time.Time `json:"created_at"`
}

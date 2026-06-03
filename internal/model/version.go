package model

type Version struct {
	BaseModel
	Version     uint32 `gorm:"column:version;uniqueIndex:idx_version_os_arch"`
	VersionName string `gorm:"column:version_name"`
	Os          string `gorm:"column:os;uniqueIndex:idx_version_os_arch"`
	Arch        string `gorm:"column:arch;uniqueIndex:idx_version_os_arch"`
	Checksum    string `gorm:"column:checksum"`
	BinarySize  int64  `gorm:"column:binary_size"`
	BinaryPath  string `gorm:"column:binary_path"`
	Changelog   string `gorm:"column:changelog;type:text"`
	FileName    string `gorm:"column:file_name"`
}

func (Version) TableName() string { return "t_version" }

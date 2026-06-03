package version

import "strconv"

// VersionStr 客户端版本号（字符串形式），编译时通过 -ldflags 注入
// VersionName 客户端版本名称，编译时通过 -ldflags 注入
// Version 由 VersionStr 转换而来的 int32 版本号
// 构建示例：
// go build -ldflags "-X github.com/Tudyha/nexus/client/version.VersionStr=2 -X github.com/Tudyha/nexus/client/version.VersionName=v2.0.0" -o nexus-cli ./main.go
var VersionStr string
var VersionName string
var Version int32

func init() {
	if v, err := strconv.Atoi(VersionStr); err == nil {
		Version = int32(v)
	}
}

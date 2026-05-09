package version

// Version 客户端版本号，编译时通过 -ldflags 注入
// VersionName 客户端版本名称，编译时通过 -ldflags 注入
// 构建示例：
// go build -ldflags "-X github.com/Tudyha/nexus/client/version.Version=2 -X github.com/Tudyha/nexus/client/version.VersionName=v2.0.0" -o nexus-cli ./main.go
var Version int32 = 1
var VersionName = "v1.0.0"

package enum

type ClientStatus uint8 // 客户端状态

const (
	ClientOnline  ClientStatus = 1 // 在线
	ClientOffline ClientStatus = 2 // 离线
)

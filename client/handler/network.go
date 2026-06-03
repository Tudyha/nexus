package handler

import (
	"strconv"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type NetworkListHandler struct{}

func NewNetworkListHandler() conn.MessageHandler      { return &NetworkListHandler{} }
func (h *NetworkListHandler) Type() proto.MessageType { return proto.MessageType_NETWORK_LIST }
func (h *NetworkListHandler) Handle(ctx conn.Context) error {
	resp, err := listNetworkConnections()
	if err != nil {
		return err
	}
	return ctx.GetConn().WriteMessage(proto.MessageType_NETWORK_LIST, resp)
}

func listNetworkConnections() (*proto.NetworkListResp, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return &proto.NetworkListResp{}, nil
	}

	var resp proto.NetworkListResp
	for _, c := range conns {
		entry := &proto.Network{
			Pid:         c.Pid,
			Protocol:    connectionType(c.Type),
			LocalAddr:   addrToString(c.Laddr),
			RemoteAddr:  addrToString(c.Raddr),
			Status:      c.Status,
			ProcessName: "",
		}

		// 尝试获取进程名
		if entry.Pid > 0 {
			if proc, err := findProcess(entry.Pid); err == nil {
				if name, err := proc.Name(); err == nil {
					entry.ProcessName = name
				}
			}
		}

		resp.List = append(resp.List, entry)
		if len(resp.List) >= 500 {
			break
		}
	}
	return &resp, nil
}

func connectionType(t uint32) string {
	switch t {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	case 3:
		return "tcp6"
	case 4:
		return "udp6"
	case 5:
		return "unix"
	default:
		return "unknown"
	}
}

func addrToString(addr net.Addr) string {
	if addr.IP == "" {
		return "0.0.0.0:0"
	}
	return addr.IP + ":" + strconv.FormatUint(uint64(addr.Port), 10)
}

func findProcess(pid int32) (*process.Process, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	return p, nil
}

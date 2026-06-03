package handler

import (
	"os"
	"syscall"
	"time"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/process"
)

type ProcessHandler struct{}

func (h *ProcessHandler) List() (*proto.ProcessListResp, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var resp proto.ProcessListResp
	for _, p := range procs {
		entry, err := fillProcessEntry(p)
		if err != nil || entry == nil {
			continue
		}
		resp.List = append(resp.List, entry)
		if len(resp.List) >= 500 {
			break
		}
	}
	return &resp, nil
}

func (h *ProcessHandler) Kill(pid int, signal int32) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.Signal(signal))
}

type ProcessListHandler struct{ ProcessHandler }

func NewProcessListHandler() conn.MessageHandler {
	return &ProcessListHandler{
		ProcessHandler: ProcessHandler{},
	}
}

func (h *ProcessListHandler) Type() proto.MessageType { return proto.MessageType_PROCESS_LIST }
func (h *ProcessListHandler) Handle(ctx conn.Context) error {
	resp, err := h.ProcessHandler.List()
	if err != nil {
		return err
	}
	return ctx.GetConn().WriteMessage(proto.MessageType_PROCESS_LIST, resp)
}

func fillProcessEntry(p *process.Process) (*proto.Process, error) {
	if p.Pid == 0 {
		return nil, nil
	}

	entry := &proto.Process{Pid: p.Pid}
	entry.Name, _ = p.Name()
	if ppid, err := p.Ppid(); err == nil {
		entry.Ppid = ppid
	}
	if threads, err := p.NumThreads(); err == nil {
		entry.Threads = threads
	}
	if username, err := p.Username(); err == nil {
		entry.User = username
	}
	if cpu, err := p.CPUPercent(); err == nil {
		entry.CpuPercent = cpu
	}
	if mem, err := p.MemoryPercent(); err == nil {
		entry.MemPercent = float64(mem)
	}
	if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
		entry.MemRss = int64(memInfo.RSS)
	}
	if status, err := p.Status(); err == nil && len(status) > 0 {
		entry.Status = status[0]
	}
	if createTime, err := p.CreateTime(); err == nil && createTime > 0 {
		entry.StartTime = timeFromUnixMillis(createTime)
	}
	return entry, nil
}

func timeFromUnixMillis(ms int64) string {
	t := time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

type ProcessKillHandler struct{ ProcessHandler }

func NewProcessKillHandler() conn.MessageHandler {
	return &ProcessKillHandler{ProcessHandler: ProcessHandler{}}
}
func (h *ProcessKillHandler) Type() proto.MessageType { return proto.MessageType_PROCESS_KILL }
func (h *ProcessKillHandler) Handle(ctx conn.Context) error {
	var req proto.ProcessKillReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	sig := req.Signal
	if sig == 0 {
		sig = 15 // 默认 SIGTERM
	}
	log.Info().Int32("pid", req.Pid).Int32("signal", sig).Msg("kill 进程")
	err := h.ProcessHandler.Kill(int(req.Pid), sig)
	if err != nil {
		log.Warn().Err(err).Int32("pid", req.Pid).Msg("kill 进程失败")
	} else {
		log.Info().Int32("pid", req.Pid).Msg("进程已 kill")
	}
	return ctx.GetConn().WriteMessage(proto.MessageType_PROCESS_KILL, &proto.Response{Code: errToCode(err)})
}

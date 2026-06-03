package handler

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"

	"syscall"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

var (
	userCache  = new(sync.Map)
	groupCache = new(sync.Map)
)

func lookupUser(uid uint32) string {
	if name, ok := userCache.Load(uid); ok {
		return name.(string)
	}
	if u, err := user.LookupId(strconv.Itoa(int(uid))); err == nil {
		userCache.Store(uid, u.Username)
		return u.Username
	}
	return strconv.Itoa(int(uid))
}

func lookupGroup(gid uint32) string {
	if name, ok := groupCache.Load(gid); ok {
		return name.(string)
	}
	if g, err := user.LookupGroupId(strconv.Itoa(int(gid))); err == nil {
		groupCache.Store(gid, g.Name)
		return g.Name
	}
	return strconv.Itoa(int(gid))
}

type FileHandler struct{}

func (h *FileHandler) List(path string) (*proto.FileListResp, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var resp proto.FileListResp
	for _, e := range entries {
		info, _ := e.Info()
		if info == nil {
			resp.Entries = append(resp.Entries, &proto.File{Name: e.Name(), IsDir: e.IsDir()})
			continue
		}
		owner, group := "", ""
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			owner = lookupUser(stat.Uid)
			group = lookupGroup(stat.Gid)
		}
		resp.Entries = append(resp.Entries, &proto.File{
			Name:    e.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
			IsDir:   e.IsDir(),
			Owner:   owner,
			Group:   group,
		})
	}
	return &resp, nil
}

func (h *FileHandler) Download(path string, offset, limit int64) (*proto.FileDownloadResp, error) {
	if containsPathTraversal(path) {
		return nil, fmt.Errorf("invalid path")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, _ := f.Stat()
	var totalSize int64
	if info != nil {
		totalSize = info.Size()
	}
	if offset > 0 {
		f.Seek(offset, io.SeekStart)
	}
	if limit <= 0 || limit > 10*1024*1024 {
		limit = 10 * 1024 * 1024
	}
	data := make([]byte, limit)
	n, _ := f.Read(data)
	return &proto.FileDownloadResp{Data: data[:n], TotalSize: totalSize}, nil
}

func (h *FileHandler) Upload(path string, data []byte, append_ bool) error {
	if containsPathTraversal(path) {
		return fmt.Errorf("invalid path")
	}
	flags := os.O_WRONLY | os.O_CREATE
	if append_ {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

type FileListHandler struct{ FileHandler }

func NewFileListHandler() conn.MessageHandler      { return &FileListHandler{FileHandler{}} }
func (h *FileListHandler) Type() proto.MessageType { return proto.MessageType_FILE_LIST }
func (h *FileListHandler) Handle(ctx conn.Context) error {
	var req proto.FileListReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	log.Debug().Str("path", req.Path).Msg("file list")
	resp, err := h.FileHandler.List(req.Path)
	if err != nil {
		return err
	}
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_LIST, resp)
}

type FileDownloadHandler struct{ FileHandler }

func NewFileDownloadHandler() conn.MessageHandler      { return &FileDownloadHandler{FileHandler{}} }
func (h *FileDownloadHandler) Type() proto.MessageType { return proto.MessageType_FILE_DOWNLOAD }
func (h *FileDownloadHandler) Handle(ctx conn.Context) error {
	var req proto.FileDownloadReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	log.Debug().Str("path", req.Path).Int64("offset", req.Offset).Int64("limit", req.Limit).Msg("file download")
	resp, err := h.FileHandler.Download(req.Path, req.Offset, req.Limit)
	if err != nil {
		return err
	}
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_DOWNLOAD, resp)
}

type FileUploadHandler struct{ FileHandler }

func NewFileUploadHandler() conn.MessageHandler      { return &FileUploadHandler{FileHandler{}} }
func (h *FileUploadHandler) Type() proto.MessageType { return proto.MessageType_FILE_UPLOAD }
func (h *FileUploadHandler) Handle(ctx conn.Context) error {
	var req proto.FileUploadReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	log.Info().Str("path", req.Path).Bool("append", req.Append).Int("size", len(req.Data)).Msg("file upload")
	err := h.FileHandler.Upload(req.Path, req.Data, req.Append)
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_UPLOAD, &proto.Response{Code: errToCode(err)})
}

type FileDeleteHandler struct{}

func NewFileDeleteHandler() conn.MessageHandler      { return &FileDeleteHandler{} }
func (h *FileDeleteHandler) Type() proto.MessageType { return proto.MessageType_FILE_DELETE }
func (h *FileDeleteHandler) Handle(ctx conn.Context) error {
	var req proto.FileDeleteReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	if containsPathTraversal(req.Path) {
		return ctx.GetConn().WriteMessage(proto.MessageType_FILE_DELETE, &proto.Response{Code: proto.ErrorCode_FAILED, Msg: "invalid path"})
	}
	log.Info().Str("path", req.Path).Msg("file delete")
	err := os.RemoveAll(req.Path)
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_DELETE, &proto.Response{Code: errToCode(err)})
}

type FileMkdirHandler struct{}

func NewFileMkdirHandler() conn.MessageHandler      { return &FileMkdirHandler{} }
func (h *FileMkdirHandler) Type() proto.MessageType { return proto.MessageType_FILE_MKDIR }
func (h *FileMkdirHandler) Handle(ctx conn.Context) error {
	var req proto.FileMkdirReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	if containsPathTraversal(req.Path) {
		return ctx.GetConn().WriteMessage(proto.MessageType_FILE_MKDIR, &proto.Response{Code: proto.ErrorCode_FAILED, Msg: "invalid path"})
	}
	log.Info().Str("path", req.Path).Msg("file mkdir")
	err := os.MkdirAll(req.Path, 0755)
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_MKDIR, &proto.Response{Code: errToCode(err)})
}

type FileRenameHandler struct{}

func NewFileRenameHandler() conn.MessageHandler      { return &FileRenameHandler{} }
func (h *FileRenameHandler) Type() proto.MessageType { return proto.MessageType_FILE_RENAME }
func (h *FileRenameHandler) Handle(ctx conn.Context) error {
	var req proto.FileRenameReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	if containsPathTraversal(req.OldPath) || containsPathTraversal(req.NewPath) {
		return ctx.GetConn().WriteMessage(proto.MessageType_FILE_RENAME, &proto.Response{Code: proto.ErrorCode_FAILED, Msg: "invalid path"})
	}
	log.Info().Str("old_path", req.OldPath).Str("new_path", req.NewPath).Msg("file rename")
	err := os.Rename(req.OldPath, req.NewPath)
	return ctx.GetConn().WriteMessage(proto.MessageType_FILE_RENAME, &proto.Response{Code: errToCode(err)})
}

func errToCode(err error) proto.ErrorCode {
	if err != nil {
		return proto.ErrorCode_FAILED
	}
	return proto.ErrorCode_SUCCESS
}

// containsPathTraversal 检查路径穿越
func containsPathTraversal(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return true
	}
	clean := filepath.Clean(abs)
	vol := filepath.VolumeName(clean)
	rest := clean[len(vol):]
	depth := 0
	for _, p := range filepath.SplitList(rest) {
		if p == ".." {
			depth--
		} else if p != "." && p != "" {
			depth++
		}
		if depth < 0 {
			return true
		}
	}
	return false
}

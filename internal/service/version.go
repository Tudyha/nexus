package service

import (
	"context"
	"errors"
	"os"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type versionService struct {
	versionDao dao.VersionDao
}

func newVersionService() VersionService {
	return &versionService{
		versionDao: dao.GetVersionDao(),
	}
}

func (s *versionService) Upload(ctx context.Context, version uint32, versionName, goos, arch, changelog, binaryPath, fileName string) error {
	existing, err := s.versionDao.GetLatestByOS(ctx, goos, arch)
	if err == nil && existing != nil && existing.Version >= version {
		return errcode.ErrVersionDuplicate
	}

	v := &model.Version{
		Version:     version,
		VersionName: versionName,
		Os:          goos,
		Arch:        arch,
		BinaryPath:  binaryPath,
		Changelog:   changelog,
		FileName:    fileName,
	}

	// 文件大小
	info, err := os.Stat(binaryPath)
	if err != nil {
		return err
	}
	v.BinarySize = info.Size()

	// 计算校验和
	if v.Checksum, err = utils.Checksum(binaryPath); err != nil {
		return err
	}

	return s.versionDao.Create(ctx, v)
}

func (s *versionService) GetPage(ctx context.Context, query request.PageQuery) (*response.Page[response.VersionResponse], error) {
	versions, total, err := s.versionDao.GetPage(ctx, query)
	if err != nil {
		return nil, err
	}
	var list []response.VersionResponse
	copier.Copy(&list, versions)
	return &response.Page[response.VersionResponse]{
		Total: total,
		List:  list,
	}, nil
}

func (s *versionService) GetByID(ctx context.Context, id uint64) (*model.Version, error) {
	return s.versionDao.GetByID(ctx, id)
}

func (s *versionService) Delete(ctx context.Context, id uint64) error {
	v, err := s.versionDao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrVersionNotFound
		}
		return err
	}

	if v.BinaryPath != "" {
		os.Remove(v.BinaryPath)
	}

	return s.versionDao.Delete(ctx, v.ID)
}

func (s *versionService) GetLatestByOS(ctx context.Context, osName, arch string) (*model.Version, error) {
	v, err := s.versionDao.GetLatestByOS(ctx, osName, arch)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrLatestVersionNotFound
		}
		return nil, err
	}
	return v, nil
}

func (s *versionService) GetLatestForClient(ctx context.Context, clientID uint64) (*model.Version, error) {
	client, err := dao.GetClientDao().GetByID(ctx, clientID)
	if err != nil {
		return nil, errcode.ErrClientNotFound
	}
	if v, err := s.versionDao.GetLatestByOS(ctx, client.Os, client.Arch); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrLatestVersionNotFound
		}
		return nil, err
	} else {
		return v, nil
	}
}

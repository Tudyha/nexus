package service

import (
	"context"
	"fmt"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/jinzhu/copier"
	"github.com/samber/lo"
)

type userService struct {
	userDao      dao.UserDao
	workSpaceDao dao.WorkspaceDao
	appDao       dao.AppDao
}

func newUserService() UserService {
	return &userService{
		userDao:      dao.GetUserDao(),
		workSpaceDao: dao.GetWorkspaceDao(),
		appDao:       dao.GetAppDao(),
	}
}

// 创建用户, 创建默认空间, 创建默认应用
func (s *userService) Create(ctx context.Context, user *model.User) error {
	if err := s.userDao.Create(ctx, user); err != nil {
		return err
	}

	user.Nickname = fmt.Sprintf("用户%06d", user.ID)
	if err := s.userDao.Update(ctx, user); err != nil {
		return err
	}

	// 创建默认空间
	space, err := s.workSpaceDao.Create(ctx, fmt.Sprintf("%s的空间", user.Nickname), "")
	if err != nil {
		return err
	}
	if err := s.workSpaceDao.CreateWorkspaceUser(ctx, space.ID, user.ID, 1); err != nil {
		return err
	}
	// 创建默认应用
	if err := s.appDao.Create(ctx, space.ID, utils.GenerateRandomString(16), "默认应用", ""); err != nil {
		return err
	}

	return nil
}

func (s *userService) Detail(ctx context.Context, id uint64) (*response.UserResponse, error) {
	user, err := s.userDao.FindByID(ctx, id)
	if err != nil {
		return nil, errcode.ErrUserNotExist
	}

	var userResp response.UserResponse
	copier.Copy(&userResp, user)
	userResp.WorkspaceList, err = s.GetSpaceList(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &userResp, nil
}

func (s *userService) GetSpaceList(ctx context.Context, userID uint64) ([]*response.WorkspaceResponse, error) {
	spaceList, err := s.workSpaceDao.GetByUserID(ctx, userID)
	if err != nil || len(spaceList) == 0 {
		return nil, err
	}

	spaceIds := lo.Map(spaceList, func(s *model.Workspace, index int) uint64 {
		return s.ID
	})
	appList, err := s.appDao.GetAppByWorkspaceIDs(ctx, spaceIds)
	if err != nil || len(appList) == 0 {
		return nil, err
	}

	appMap := lo.GroupByMap(appList, func(item *model.App) (uint64, *model.App) {
		return item.WorkspaceID, item
	})

	var list []*response.WorkspaceResponse

	copier.Copy(&list, spaceList)

	for _, space := range list {
		apps := appMap[space.ID]
		if len(apps) > 0 {
			var appResp []*response.AppResponse
			copier.Copy(&appResp, apps)
			space.AppList = appResp
		}
	}

	return list, nil
}

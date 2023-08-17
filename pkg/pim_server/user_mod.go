package pim_server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"pim/api"
	"pim/pkg/models"
	"pim/pkg/tools"
	"time"
)

func (p *PimServer) Register(ctx context.Context, req *api.RegisterReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	// 这个接口没有鉴权

	db := p.svr.db
	logger := p.svr.logger
	var authInfo models.Auth

	authInfo.Password = req.Password
	authInfo.Email = req.Email
	authInfo.Username = req.Username

	errSave := db.Create(&authInfo).Error
	if errSave != nil {

		logger.Error("注册账户失败", zap.Any("params", req), zap.Error(errSave))
		err = errors.New("注册失败")
		return
	}

	resp = new(api.BaseOk)
	//// 创建用户信息
	//var userinfo models.UserInfo
	//
	//userinfo.UserID = authInfo.UserID
	//userinfo.Username = req.Username
	//
	//err = s.db.Create(&userinfo).Error
	//
	//if err != nil {
	//	tools.Resp500(ctx, err.Error())
	//	return
	//}

	return

}

func (p *PimServer) Login(ctx context.Context, req *api.LoginReq) (resp *api.LoginResp, err error) {
	//TODO implement me
	//panic("implement me")
	// 查找 用户

	db := p.svr.db

	logger := p.svr.logger

	switch req.Type {
	case api.LoginReq_LoginByUsername:
	default:
		err = errors.New("不支持的登录类型")
		return
	}
	var userinfo models.Auth

	errFind := db.
		Model(&userinfo).
		Where("(email = ? or username = ? or mobile = ?) and password = ? ",
			req.Username, req.Username, req.Username, req.Password).
		Find(&userinfo).Error

	if errFind != nil || userinfo.UserID == 0 {
		logger.Error("查询账户失败", zap.Error(errFind))
		err = errors.New("登录失败")
		return
	}

	tokenStr, _ := tools.GenToken(userinfo.UserID, int(req.Platform), userinfo.Level)

	resp = new(api.LoginResp)
	resp.Token = tokenStr
	return

}

func (p *PimServer) GetMyUserInfo(ctx context.Context, req *api.StreamReq) (resp *api.UserInfoViewerDataType, err error) {
	// 从流中提取基本信息

	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 查询我的信息

	logger := p.svr.logger
	db := p.svr.db

	var userinfo models.UserInfoViewer

	err = db.Model(&userinfo).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Find(&userinfo).Error

	if err != nil || userinfo.UserID == 0 {
		logger.Debug("get user info by id fail ", zap.Error(err))
		err = errors.New("user not found")
		return
	}
	resp = new(api.UserInfoViewerDataType)
	resp.UserID = userinfo.UserID
	resp.Username = userinfo.Username
	resp.Email = userinfo.Email
	resp.Nick = userinfo.Nick
	resp.CreatedAt = userinfo.CreatedAt
	resp.Avatar = userinfo.Avatar
	resp.UserType = api.UserTypeEnumType(userinfo.UserType)
	resp.UserStatus = api.UserStatusEnumType(userinfo.UserStatus)

	// avatar ...

	return
}

func (p *PimServer) GetUserInfoByID(ctx context.Context, req *api.GetUserInfoByIDReq) (resp *api.UserInfoViewerDataType, err error) {
	//TODO implement me
	//panic("implement me")
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 查询目标用户的

	logger := p.svr.logger
	db := p.svr.db

	var userinfo models.UserInfoViewer

	err = db.Model(&userinfo).Where(&models.UserInfoViewer{
		UserID: req.UserID,
	}).Find(&userinfo).Error

	if err != nil || userinfo.UserID == 0 {
		logger.Debug("get user info by id fail ", zap.Error(err))
		err = errors.New("user not found")
		return
	}
	resp = new(api.UserInfoViewerDataType)
	resp.UserID = userinfo.UserID
	resp.Username = userinfo.Username
	resp.Email = userinfo.Email
	resp.Nick = userinfo.Nick
	resp.CreatedAt = userinfo.CreatedAt
	resp.Avatar = userinfo.Avatar
	resp.UserType = api.UserTypeEnumType(userinfo.UserType)
	resp.UserStatus = api.UserStatusEnumType(userinfo.UserStatus)

	// avatar ...

	return
}

func (p *PimServer) AddUserToContact(ctx context.Context, req *api.AddUserToContactReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	//panic("implement me")
	resp = new(api.BaseOk)

	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo

	// 数据加入数据库

	req.UserID = tokenInfo.GetUserID()
	//

	db := p.svr.db
	logger := p.svr.logger
	// 判断用户是否存在

	var findUser models.UserInfoViewer

	//findUser.UserID = cUserID

	respDB := db.Model(&findUser).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Find(&findUser)

	if respDB.Error != nil {
		//err = respDB.Error
		logger.Debug("查找用户失败", zap.Error(respDB.Error))
		err = errors.New("添加失败")
		return
	}

	updateMap := map[string]interface{}{}

	//if req.FirstName != "" {
	//	updateMap["first_name"] = req.FirstName
	//}
	//
	//if req.LastName != "" {
	//	updateMap["last_name"] = req.LastName
	//}

	if req.Note != "" {
		updateMap["note"] = req.Note
	}

	timeNow := time.Now()

	updateMap["updated_at"] = timeNow.UnixMilli()

	// 可以添加
	addErr := db.Clauses(&clause.OnConflict{
		//DoNothing: false,
		DoUpdates: clause.Assignments(updateMap),
	}).Create(req).Error

	if addErr != nil {
		logger.Debug("添加用户失败", zap.Error(addErr))

		err = errors.New("添加失败")
		return
	}

	return
}

// UserUpdateInfo 更新用户信息
func (p *PimServer) UserUpdateUsername(ctx context.Context, req *api.UserUpdateInfoReq) (resp *api.BaseOk, err error) {

	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}

	// 判断需要更新的东西

	var updateUserInfo models.UserInfoViewer

	if req.Username == nil {
		// 更新用户名
		// 检验用户名是否
		//updateUserInfo.Username

		// 用户名参数错误
		err = errors.New("参数错误")
		return
	}

	// 查看用户名是否被使用

	var findUserName models.UserInfoViewer

	db := p.svr.db

	logger := p.svr.logger
	findErr := db.Model(&findUserName).Where(&models.UserInfoViewer{
		Username: req.GetUsername(),
	}).Find(&findUserName).Error
	if findErr != nil {
		logger.Debug("数据库错误", zap.Error(err))
		err = errors.New("数据库错误")
		return
	}

	if findUserName.UserID != 0 {
		err = errors.New("用户名已被使用")
		return
	}

	// 可以使用的用户名

	updateErr := db.Model(&updateUserInfo).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Update("username", req.GetUsername()).Error

	if updateErr != nil {
		logger.Debug("update username fail", zap.Error(updateErr))
		err = errors.New("更新username fail")
		return
	}
	resp = new(api.BaseOk)

	//清理缓存

	_ = p.svr.dao.DeleteUserInfoCache(tokenInfo.GetUserID())

	return
}

func (p *PimServer) UserUpdateNick(ctx context.Context, req *api.UserUpdateInfoReq) (resp *api.BaseOk, err error) {
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}

	// 判断需要更新的东西

	var updateUserInfo models.UserInfoViewer

	if req.Nick == nil {
		// 更新用户名
		// 检验用户名是否
		//updateUserInfo.Username

		// 用户名参数错误
		err = errors.New("参数错误")
		return
	}

	// 查看用户名是否被使用

	db := p.svr.db

	logger := p.svr.logger

	// 可以使用的用户名

	updateErr := db.Model(&updateUserInfo).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Update("nick", req.GetNick()).Error

	if updateErr != nil {
		logger.Debug("update nick fail", zap.Error(updateErr))
		err = errors.New("更新 nick fail")
		return
	}
	resp = new(api.BaseOk)
	_ = p.svr.dao.DeleteUserInfoCache(tokenInfo.GetUserID())

	return
}

func (p *PimServer) UserUpdateAvatar(ctx context.Context, req *api.UserUpdateInfoReq) (resp *api.BaseOk, err error) {
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}

	// 判断需要更新的东西

	var updateUserInfo models.UserInfoViewer

	if req.Avatar == nil {
		// 更新用户名
		// 检验用户名是否
		//updateUserInfo.Username

		// 用户名参数错误
		err = errors.New("参数错误")
		return
	}

	// 查看用户名是否被使用

	db := p.svr.db

	logger := p.svr.logger

	// 可以使用的用户名

	updateErr := db.Model(&updateUserInfo).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Update("avatar", req.GetAvatar()).Error

	if updateErr != nil {
		logger.Debug("update avatar fail", zap.Error(updateErr))
		err = errors.New("更新 avatar fail")
		return
	}
	resp = new(api.BaseOk)
	_ = p.svr.dao.DeleteUserInfoCache(tokenInfo.GetUserID())

	return
}

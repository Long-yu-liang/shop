package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"shop_servs/user_srv/global"
	"shop_servs/user_srv/model"
	proto "shop_servs/user_srv/proto/user"
	"strings"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

//	type UserServer interface {
//		GetUserList(context.Context, *PageInfo) (*UserListResponse, error)
//		GetUserByMobile(context.Context, *MobileRequest) (*UserInfoResonse, error)
//		GetUserById(context.Context, *IdRequest) (*UserInfoResonse, error)
//		CreateUser(context.Context, *CreateUserInfo) (*UserInfoResonse, error)
//		UpdateUser(context.Context, *UpdateUserInfo) (*emptypb.Empty, error)
//		CheckPassword(context.Context, *PasswordCheckInfo) (*CheckResponse, error)
//		mustEmbedUnimplementedUserServer()-->proto.UnimplementedUserServer
//	}
type UserServer struct {
	proto.UnimplementedUserServer
}

// 将model.User(结构体) 转换成proto.UserInfoResonse
func ModelToResponse(user model.User) *proto.UserInfoResonse {
	//在grpc中message中字段有默认值，不能随便赋值nil进去，容易出错
	//搞清楚那些字段有默认值
	userInfoRsp := &proto.UserInfoResonse{
		Id:       user.ID,
		Password: user.Password,
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}
	return userInfoRsp
}

// 分页
func Paginate(page, pageSize int, orderBy ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 设置默认值
		if page < 1 {
			page = 1
		}

		switch {
		case pageSize < 1:
			pageSize = 10 // 默认值
		case pageSize > 100: // 防止过度查询，设置最大限制
			pageSize = 100
		}

		offset := (page - 1) * pageSize

		// 应用分页
		db = db.Offset(offset).Limit(pageSize)

		// 如果有排序参数，应用排序
		if len(orderBy) > 0 && orderBy[0] != "" {
			db = db.Order(orderBy[0])
		}

		return db
	}
}

// 获得用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	var users []model.User
	var total int64

	// 获取总数
	global.DB.Model(&model.User{}).Count(&total)

	// 获取分页数据
	result := global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	//rsp为返回的UserListResponse表
	rsp := &proto.UserListResponse{}
	rsp.Total = int32(total)
	//rsp.Total = int32(result.RowsAffected) // 同样记录rsp.Total

	for _, user := range users {
		userInfoRsp := ModelToResponse(user)
		rsp.Data = append(rsp.Data, userInfoRsp)
	}
	return rsp, nil
}

// 通过手机号码查询用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResonse, error) {
	var user model.User
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	userInfoRsp := ModelToResponse(user)
	return userInfoRsp, nil
}

// 通过id查询用户
func (s *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResonse, error) {
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	userInfoRsp := ModelToResponse(user)
	return userInfoRsp, nil
}

// 创建用户
func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResonse, error) {
	//新建用户
	var user model.User
	//查询是否已经存在
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}

	user.Mobile = req.Mobile
	user.NickName = req.NickName
	//密码加密
	options := &password.Options{
		SaltLen:      16,
		Iterations:   100,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	salt, encodePwd := password.Encode(req.Password, options)
	newpassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodePwd)
	user.Password = newpassword

	//完成user数据后，在数据库创建user
	result = global.DB.Create(&user)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}

	userInfoRsp := ModelToResponse(user)

	return userInfoRsp, nil
}

// 更新用户数据信息
func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	//查询用户是否已经存在
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	//
	birthDay := time.Unix(int64(req.Birthday), 0)
	user.NickName = req.NickName
	//user.Birthday 为 *time.Time 我们要将 proto 的int64的 birthday 转为 *time.Time
	user.Birthday = &birthDay
	user.Gender = req.Gender
	result = global.DB.Save(&user)
	if result.Error != nil {

	}
	return &emptypb.Empty{}, nil
}

// 校验密码
func (s *UserServer) CheckPassword(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	options := &password.Options{
		SaltLen:      16,
		Iterations:   100,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	passwordInfo := strings.Split(req.EncryptedPassword, "$")
	check := password.Verify(req.Password, passwordInfo[2], passwordInfo[3], options)
	return &proto.CheckResponse{
		Success: check,
	}, nil
}

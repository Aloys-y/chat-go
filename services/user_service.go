package services

import (
	"context"
	"errors"

	"github.com/Aloys-y/chat-go/auth"
	"github.com/Aloys-y/chat-go/db"
	"github.com/Aloys-y/chat-go/models"
	"github.com/Aloys-y/chat-go/proto"
)

type UserServiceImpl struct {
	proto.UnimplementedUserServiceServer
}

// Register implements UserServiceServer
func (s *UserServiceImpl) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// 参数验证
	if req.Username == "" || req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return nil, errors.New("username, email, password, and display name are required")
	}

	// 1. 调用认证模块的注册函数
	user, err := auth.RegisterUser(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		return nil, err
	}

	// 2. 生成JWT令牌
	token, err := auth.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// 3. 构造并返回响应
	return &proto.RegisterResponse{
		User: &proto.UserInfo{
			Id:          uint64(user.ID),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			IsOnline:    user.IsOnline,
		},
		Token: token,
	}, nil
}

// Login implements UserServiceServer
func (s *UserServiceImpl) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, token, err := auth.LoginUser(req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		User: &proto.UserInfo{
			Id:          uint64(user.ID),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			IsOnline:    user.IsOnline,
		},
		Token: token,
	}, nil
}

// GetUserInfo implements UserServiceServer
func (s *UserServiceImpl) GetUserInfo(ctx context.Context, req *proto.GetUserInfoRequest) (*proto.UserInfo, error) {
	user, err := auth.GetUserByID(uint(req.UserId))
	if err != nil {
		return nil, err
	}

	return &proto.UserInfo{
		Id:          uint64(user.ID),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		IsOnline:    user.IsOnline,
	}, nil
}

// UpdateUserStatus implements UserServiceServer
func (s *UserServiceImpl) UpdateUserStatus(ctx context.Context, req *proto.UpdateUserStatusRequest) (*proto.Empty, error) {
	var user models.User
	if err := db.DB.First(&user, req.UserId).Error; err != nil {
		return nil, err
	}

	user.IsOnline = req.IsOnline
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 定义context键常量，避免拼写错误
const (
	CtxUserID = "user_id"
	CtxEmail  = "email"
)

// AuthInterceptor 是一个gRPC拦截器，用于验证JWT token
type AuthInterceptor struct{}

// NewAuthInterceptor 创建一个新的认证拦截器
func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// UnaryInterceptor 实现grpc.UnaryServerInterceptor接口
func (interceptor *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 跳过一些不需要认证的方法，如注册和登录
	if info.FullMethod == "/chat.UserService/Register" || info.FullMethod == "/chat.UserService/Login" {
		return handler(ctx, req)
	}

	// 从metadata中获取token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	// 检查是否有Authorization头
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	}

	// 提取token（去掉"Bearer "前缀）
	authHeader := values[0]
	tokenString := ""
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	} else {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header format must be Bearer {token}")
	}

	// 验证token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token: %v", err)
	}

	// 将用户信息添加到context中，以便后续处理
	ctx = context.WithValue(ctx, CtxUserID, claims.UserID)
	ctx = context.WithValue(ctx, CtxEmail, claims.Email)

	// 调用实际的处理函数
	return handler(ctx, req)
}

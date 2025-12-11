# Chat-Go

Chat-Go是一个基于Go语言的实时语音聊天应用，使用gRPC、WebSocket和WebRTC技术栈实现。

## 功能特性

- 用户认证（注册、登录）
- 房间管理（创建、加入、离开房间）
- 实时语音通信（基于WebRTC）
- 信令服务器（基于WebSocket）
- 多房间支持
- 在线用户状态显示

## 技术栈

- **后端**：Go 1.21+
- **数据库**：MySQL
- **API**：gRPC
- **实时通信**：WebSocket + WebRTC
- **ORM**：GORM
- **配置管理**：Viper
- **认证**：JWT

## 项目结构

```
chat-go/
├── auth/          # 认证相关功能
├── config/        # 配置管理
├── db/            # 数据库连接
├── models/        # 数据模型
├── proto/         # gRPC协议定义
├── services/      # gRPC服务实现
├── signaling/     # WebSocket信令服务
├── web/           # Web客户端
├── go.mod         # Go模块依赖
├── main.go        # 项目入口
└── README.md      # 项目说明
```

## 安装与运行

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置MySQL数据库

在`config/config.yaml`文件中配置数据库连接信息：

```yaml
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_password"
  name: "char_go"
  charset: "utf8mb4"
  parseTime: true
```

### 3. 生成gRPC代码

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/improbable-eng/grpc-web/go/grpcwebproxy/protoc-gen-grpc-web@v0.15.0

protoc --go_out=. --go-grpc_out=. proto/char.proto
protoc --grpc-web_out=import_style=commonjs,mode=grpcwebtext:. proto/char.proto
```

### 4. 运行项目

```bash
go run main.go
```

### 5. 访问Web客户端

在浏览器中打开：`http://localhost:8080`

## API文档

### gRPC服务

#### UserService
- `Register` - 用户注册
- `Login` - 用户登录
- `GetUserInfo` - 获取用户信息
- `UpdateUserStatus` - 更新用户状态

#### RoomService
- `CreateRoom` - 创建房间
- `JoinRoom` - 加入房间
- `LeaveRoom` - 离开房间
- `GetRoomInfo` - 获取房间信息
- `ListRooms` - 列出所有房间
- `ListRoomUsers` - 列出房间内用户

### WebSocket消息类型

#### 客户端发送
- `join_room` - 加入房间
- `leave_room` - 离开房间
- `sdp_offer` - WebRTC SDP offer
- `sdp_answer` - WebRTC SDP answer
- `ice_candidate` - WebRTC ICE候选

#### 服务器发送
- `user_joined` - 用户加入房间通知
- `user_left` - 用户离开房间通知
- `sdp_offer` - 转发WebRTC SDP offer
- `sdp_answer` - 转发WebRTC SDP answer
- `ice_candidate` - 转发WebRTC ICE候选

## 开发说明

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Node.js 14+ (用于生成gRPC Web代码)

### 测试

```bash
go test ./...
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## TODO

- [ ] 添加视频通信功能
- [ ] 添加消息聊天功能
- [ ] 优化WebRTC连接稳定性
- [ ] 添加移动端客户端
- [ ] 实现房间密码保护
- [ ] 添加用户禁言功能
- [ ] 实现录音功能
- [ ] 添加房间管理后台
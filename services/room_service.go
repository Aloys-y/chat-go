package services

import (
	"context"
	"errors"

	"github.com/yourusername/chat-go/db"
	"github.com/yourusername/chat-go/models"
	"github.com/yourusername/chat-go/proto"
)

type RoomServiceImpl struct{}

// CreateRoom implements RoomServiceServer
func (s *RoomServiceImpl) CreateRoom(ctx context.Context, req *proto.CreateRoomRequest) (*proto.RoomInfo, error) {
	var owner models.User
	if err := db.DB.First(&owner, req.OwnerId).Error; err != nil {
		return nil, err
	}

	room := &models.Room{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		OwnerID:     uint(req.OwnerId),
	}

	if err := db.DB.Create(room).Error; err != nil {
		return nil, err
	}

	// Add owner to the room
	if err := db.DB.Model(room).Association("Users").Append(&owner); err != nil {
		return nil, err
	}

	return &proto.RoomInfo{
		Id:          uint64(room.ID),
		Name:        room.Name,
		Description: room.Description,
		IsPublic:    room.IsPublic,
		Owner: &proto.UserInfo{
			Id:          uint64(owner.ID),
			Username:    owner.Username,
			Email:       owner.Email,
			DisplayName: owner.DisplayName,
			IsOnline:    owner.IsOnline,
		},
		UserCount: 1,
	}, nil
}

// JoinRoom implements RoomServiceServer
func (s *RoomServiceImpl) JoinRoom(ctx context.Context, req *proto.JoinRoomRequest) (*proto.RoomInfo, error) {
	var room models.Room
	var user models.User

	if err := db.DB.First(&room, req.RoomId).Error; err != nil {
		return nil, err
	}

	if err := db.DB.First(&user, req.UserId).Error; err != nil {
		return nil, err
	}

	// Check if user is already in the room
	var count int
	if err := db.DB.Model(&room).Where("id = ?", req.UserId).Association("Users").Count(&count); err != nil || count > 0 {
		return nil, nil // User already in room
	}

	// Add user to the room
	if err := db.DB.Model(&room).Association("Users").Append(&user); err != nil {
		return nil, err
	}

	// Get updated room info
	if err := db.DB.Preload("Owner").First(&room, req.RoomId).Error; err != nil {
		return nil, err
	}

	return &proto.RoomInfo{
		Id:          uint64(room.ID),
		Name:        room.Name,
		Description: room.Description,
		IsPublic:    room.IsPublic,
		Owner: &proto.UserInfo{
			Id:          uint64(room.Owner.ID),
			Username:    room.Owner.Username,
			Email:       room.Owner.Email,
			DisplayName: room.Owner.DisplayName,
			IsOnline:    room.Owner.IsOnline,
		},
		UserCount: int32(db.DB.Model(&room).Association("Users").Count()),
	}, nil
}

// LeaveRoom implements RoomServiceServer
func (s *RoomServiceImpl) LeaveRoom(ctx context.Context, req *proto.LeaveRoomRequest) (*proto.Empty, error) {
	var room models.Room
	var user models.User

	if err := db.DB.First(&room, req.RoomId).Error; err != nil {
		return nil, err
	}

	if err := db.DB.First(&user, req.UserId).Error; err != nil {
		return nil, err
	}

	// Remove user from the room
	if err := db.DB.Model(&room).Association("Users").Delete(&user); err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

// GetRoomInfo implements RoomServiceServer
func (s *RoomServiceImpl) GetRoomInfo(ctx context.Context, req *proto.GetRoomInfoRequest) (*proto.RoomInfo, error) {
	var room models.Room
	if err := db.DB.Preload("Owner").First(&room, req.RoomId).Error; err != nil {
		return nil, err
	}

	return &proto.RoomInfo{
		Id:          uint64(room.ID),
		Name:        room.Name,
		Description: room.Description,
		IsPublic:    room.IsPublic,
		Owner: &proto.UserInfo{
			Id:          uint64(room.Owner.ID),
			Username:    room.Owner.Username,
			Email:       room.Owner.Email,
			DisplayName: room.Owner.DisplayName,
			IsOnline:    room.Owner.IsOnline,
		},
		UserCount: int32(db.DB.Model(&room).Association("Users").Count()),
	}, nil
}

// ListRooms implements RoomServiceServer
func (s *RoomServiceImpl) ListRooms(ctx context.Context, req *proto.ListRoomsRequest) (*proto.ListRoomsResponse, error) {
	var rooms []models.Room
	var totalCount int64

	query := db.DB.Model(&models.Room{}).Where("is_public = ?", req.IsPublic)
	query.Count(&totalCount)

	// Pagination
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(int(offset)).Limit(int(req.PageSize)).Preload("Owner").Find(&rooms).Error; err != nil {
		return nil, err
	}

	roomInfos := make([]*proto.RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		userCount := int32(db.DB.Model(&room).Association("Users").Count())
		roomInfos = append(roomInfos, &proto.RoomInfo{
			Id:          uint64(room.ID),
			Name:        room.Name,
			Description: room.Description,
			IsPublic:    room.IsPublic,
			Owner: &proto.UserInfo{
				Id:          uint64(room.Owner.ID),
				Username:    room.Owner.Username,
				Email:       room.Owner.Email,
				DisplayName: room.Owner.DisplayName,
				IsOnline:    room.Owner.IsOnline,
			},
			UserCount: userCount,
		})
	}

	return &proto.ListRoomsResponse{
		Rooms:      roomInfos,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// ListRoomUsers implements RoomServiceServer
func (s *RoomServiceImpl) ListRoomUsers(ctx context.Context, req *proto.ListRoomUsersRequest) (*proto.ListUsersResponse, error) {
	var room models.Room
	if err := db.DB.First(&room, req.RoomId).Error; err != nil {
		return nil, err
	}

	var users []models.User
	if err := db.DB.Model(&room).Association("Users").Find(&users).Error; err != nil {
		return nil, err
	}

	userInfos := make([]*proto.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, &proto.UserInfo{
			Id:          uint64(user.ID),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			IsOnline:    user.IsOnline,
		})
	}

	return &proto.ListUsersResponse{Users: userInfos}, nil
}
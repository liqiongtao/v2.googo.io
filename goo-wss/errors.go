package goowss

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("websocket client not found")
	// ErrRoomNotFound 房间未找到
	ErrRoomNotFound = errors.New("room not found")
	// ErrClientAlreadyInRoom 客户端已在房间中
	ErrClientAlreadyInRoom = errors.New("client already in room")
	// ErrClientNotInRoom 客户端不在房间中
	ErrClientNotInRoom = errors.New("client not in room")
	// ErrConnectionClosed 连接已关闭
	ErrConnectionClosed = errors.New("connection closed")
)


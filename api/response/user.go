package response

import user2 "myGinServer/models/user"

// UserListResp 用户分页列表响应体。
type UserListResp struct {
	List  []user2.User `json:"list"`
	Total int64        `json:"total"`
}

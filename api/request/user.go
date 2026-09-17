package request

// CreateUserReq 管理员创建用户请求体。
type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	PhoneNum string `json:"phone_num"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserReq 管理员更新用户请求体（不含密码与状态）。
type UpdateUserReq struct {
	Email    string `json:"email"`
	PhoneNum string `json:"phone_num"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// UpdateProfileReq 当前用户更新个人资料请求体。
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	PhoneNum string `json:"phone_num"`
}

// ChangePasswordReq 修改当前用户密码请求体。
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPasswordReq 管理员重置用户密码请求体。
type ResetPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required"`
}

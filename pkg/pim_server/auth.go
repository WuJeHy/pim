package pim_server

// TokenInfo 令牌信息
type TokenInfo interface {
	GetUserID() int64
	GetPf() int
	GetLevel() int
}

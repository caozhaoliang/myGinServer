package article

import (
	"fmt"
	"time"
)

// 定义自定义时间类型，嵌入time.Time
type CustomTime time.Time

// 实现json.Marshaler接口，自定义JSON序列化格式
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	// 转换为time.Time类型
	t := time.Time(ct)
	// 定义输出格式，例如："2006-01-02 15:04:05"
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

type ArticleVO struct {
	Id           string     `json:"id" db:"id"`
	Title        string     `json:"title" db:"title"`
	Cover        string     `json:"cover" db:"cover"`
	Status       string     `json:"status" db:"status"`
	ChannelId    string     `json:"channel_id" db:"channel_id"`
	Pubdate      CustomTime `json:"pubdate" db:"pubdate"`
	ViewCount    int        `json:"view_count" db:"view_count"`
	CommentCount int        `json:"comment_count" db:"comment_count"`
	LikeCount    int        `json:"like_count" db:"like_count"`
}

type ArticlesRequest struct {
	Page         int    `json:"page" form:"page"`
	PageSize     int    `json:"page_size" form:"page_size"`
	Status       string `json:"status" form:"status"`
	ChannelId    string `json:"channel_id" form:"channel_id"`
	PubdateStart string `json:"pubdate_start" form:"pubdate_start"`
	PubdateEnd   string `json:"pubdate_end" form:"pubdate_end"`
}

type ArticlesResponse struct {
	Data  []ArticleVO `json:"data"`
	Total int         `json:"total"`
}

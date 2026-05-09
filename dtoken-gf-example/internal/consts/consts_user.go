// @Author daixk 2026/5/9 22:56:00
package consts

import "github.com/gogf/gf/v2/container/gmap"

// UserData demo user data 示例用户数据
type UserData struct {
	Id       string // User id 用户 ID
	UserName string // Login user name 登录用户名
	Password string // Login password 登录密码
}

// UserDataMap demo user map, user name as key 示例用户映射，用户名为 key
var UserDataMap = gmap.NewStrAnyMapFrom(map[string]interface{}{
	"test1": UserData{
		Id:       "test1",
		UserName: "test1",
		Password: "test1",
	},
	"test2": UserData{
		Id:       "test2",
		UserName: "test2",
		Password: "test2",
	},
}, true)

package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID               uint            `gorm:"primaryKey;autoIncrement;comment:ID" json:"id"`
	GroupID          uint            `gorm:"column:group_id;not null;default:0;comment:组别ID" json:"group_id"`
	Username         string          `gorm:"column:username;type:varchar(32);default:'';comment:用户名" json:"username"`
	NumberCard       *string         `gorm:"column:numbercard;type:varchar(100);unique;comment:身份证号" json:"numbercard"`
	Password         string          `gorm:"column:password;type:varchar(32);default:'';comment:密码" json:"password"`
	Pwd              *string         `gorm:"column:pwd;type:varchar(32)" json:"pwd"`
	Mobile           string          `gorm:"column:mobile;type:varchar(11);unique;default:'';comment:手机号" json:"mobile"`
	Level            uint8           `gorm:"column:level;not null;default:0;comment:等级" json:"level"`
	PrevTime         *int64          `gorm:"column:prevtime;comment:上次登录时间" json:"prevtime"`
	LoginTime        int64           `gorm:"primaryKey;column:logintime;not null;comment:登录时间" json:"logintime"`
	LoginIP          string          `gorm:"column:loginip;type:varchar(50);default:'';comment:登录IP" json:"loginip"`
	LoginFailure     uint8           `gorm:"column:loginfailure;not null;default:0;comment:失败次数" json:"loginfailure"`
	LoginFailureTime *int64          `gorm:"column:loginfailuretime;comment:最后登录失败时间" json:"loginfailuretime"`
	JoinIP           string          `gorm:"column:joinip;type:varchar(50);default:'';comment:加入IP" json:"joinip"`
	JoinTime         *int64          `gorm:"column:jointime;comment:加入时间" json:"jointime"`
	CreateTime       *int64          `gorm:"column:createtime;comment:创建时间" json:"createtime"`
	UpdateTime       *int64          `gorm:"column:updatetime;comment:更新时间" json:"updatetime"`
	Token            string          `gorm:"column:token;type:varchar(50);default:'';comment:Token" json:"token"`
	Status           string          `gorm:"column:status;type:varchar(30);default:'';comment:状态" json:"status"`
	SuperiorsID      int             `gorm:"column:superiors_id;default:0;comment:上级ID" json:"superiors_id"`
	RecommendedMoney decimal.Decimal `gorm:"column:recommended_money;type:decimal(10,2);default:0.00;comment:推荐奖励" json:"recommended_money"`
	IsReal           string          `gorm:"column:is_real;type:enum('0','1');default:'0';comment:是否认真:0=未认真,1=已认证" json:"is_real"`
	Code             *string         `gorm:"column:code;type:varchar(255);unique;comment:邀请码" json:"code"`
	Num1             int             `gorm:"column:num1;default:0" json:"num1"`
	Num2             int             `gorm:"column:num2;default:0" json:"num2"`
	Num3             int             `gorm:"column:num3;default:0" json:"num3"`
	UseApp           *string         `gorm:"column:use_app;type:varchar(100)" json:"use_app"`
	SignMoney        decimal.Decimal `gorm:"column:sign_money;type:decimal(10,2);not null;default:0.00" json:"sign_money"`
	WithdrawLimit    int             `gorm:"column:withdraw_limit;default:0" json:"withdraw_limit"`
	RealTime         *time.Time      `gorm:"column:real_time" json:"real_time"`
}

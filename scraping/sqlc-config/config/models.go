// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package config

import (
	"database/sql"
	"time"
)

type Config struct {
	ID int32
	// 开始区块
	StartBlock int32
	// 请求接口
	Url string
	// 采集状态
	Status bool
	// 首笔转账金额条件
	FirstAmount string
	// 第二笔转账金额条件
	SecondAmount string
	// 第二笔转账时间内
	TimeLimit int32
	// 金额条件
	AmountCondition string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
	// 小飞机机器人接口地址
	TgBotUrl string
	// 小飞机群id
	TgID string
}
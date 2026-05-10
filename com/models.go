package com

import "time"

// ApiLog 모든 HTTP 요청 로그. ApiLogging 미들웨어가 자동 저장.
// RegDttm gorm:"->" 는 DB DEFAULT NOW() 위임 — WAS가 현재 시각을 직접 세팅하지 않는다.
type ApiLog struct {
	ID          int64   `gorm:"primaryKey;autoIncrement"`
	MemberID    *string `gorm:"type:varchar(128)"`
	Method      string  `gorm:"type:varchar(10);not null"`
	Endpoint    string  `gorm:"type:varchar(200);not null"`
	StatusCode  int     `gorm:"not null"`
	RecordCount *int
	LatencyMs   int       `gorm:"not null"`
	RegDttm     time.Time `gorm:"->"`
}

func (ApiLog) TableName() string { return "api_log" }

package com

import (
	"context"
	"database/sql"
	_ "embed"
	"time"
)

//go:embed time_dao.sql
var timeDaoSQL string

var timeQ = ParseNamedQueries(timeDaoSQL)

// Now DB 측 현재 시각을 반환한다. WAS 다중 인스턴스 시계 드리프트 방지용.
// service/handler에서 "현재 시각" 변수가 필요할 때 time.Now() 대신 사용한다.
func Now(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	sqlText, _, err := timeQ.MustGet("now").Render(map[string]any{})
	if err != nil {
		return time.Time{}, err
	}
	var now time.Time
	return now, tx.QueryRowContext(ctx, sqlText).Scan(&now)
}

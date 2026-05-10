package com

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// SQLTx GORM 트랜잭션 안에서 동일 *sql.Tx를 추출해 raw SQL DAO에서 공유한다.
// service의 db.Transaction(...) 내부에서만 호출하는 것이 전제.
func SQLTx(tx *gorm.DB) (*sql.Tx, error) {
	sqlTx, ok := tx.Statement.ConnPool.(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("GORM ConnPool이 *sql.Tx 아님 — 트랜잭션 외부에서 호출됐는지 확인")
	}
	return sqlTx, nil
}

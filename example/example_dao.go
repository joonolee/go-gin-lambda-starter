package example

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"time"

	"go-gin-lambda-starter/com"
)

//go:embed example_dao.sql
var exampleDaoSQL string

// exampleQ 패키지 초기화 시 SQL 파싱 완료 — MustGet은 쿼리명 오타 시 패닉
var exampleQ = com.ParseNamedQueries(exampleDaoSQL)

// Item 도메인 모델 (전송용 + DB 행 매핑 모두 사용).
type Item struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	RegDttm time.Time `json:"regDttm"`
}

// SelectItemByID 단건 조회. 없으면 sql.ErrNoRows.
func SelectItemByID(ctx context.Context, tx *sql.Tx, memberID string, id int64) (*Item, error) {
	sqlText, args, err := exampleQ.MustGet("select_item_by_id").Render(map[string]any{
		"id":        id,
		"member_id": memberID,
	})
	if err != nil {
		return nil, err
	}
	var it Item
	err = tx.QueryRowContext(ctx, sqlText, args...).Scan(&it.ID, &it.Name, &it.RegDttm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &it, nil
}

// InsertItem 신규 행 INSERT 후 id 반환.
func InsertItem(ctx context.Context, tx *sql.Tx, memberID, name, regPgm string) (int64, error) {
	sqlText, args, err := exampleQ.MustGet("insert_item_returning_id").Render(map[string]any{
		"member_id": memberID,
		"name":      name,
		"reg_pgm":   regPgm,
	})
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, sqlText, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// SelectItemsByMember 회원의 모든 아이템 (생성 역순).
func SelectItemsByMember(ctx context.Context, tx *sql.Tx, memberID string) ([]Item, error) {
	sqlText, args, err := exampleQ.MustGet("select_items_by_member").Render(map[string]any{
		"member_id": memberID,
	})
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Name, &it.RegDttm); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

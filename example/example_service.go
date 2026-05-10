package example

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"go-gin-lambda-starter/com"
)

// ErrNotFound 도메인 레벨 not-found. handler에서 404 매핑에 사용.
var ErrNotFound = errors.New("item not found")

// Metadata handler가 요청 컨텍스트에서 추출해 내려보내는 공통 메타.
type Metadata struct {
	MemberID string // FirebaseAuth 미들웨어가 컨텍스트에 세팅한 sub
	RegPgm   string // 호출된 엔드포인트 경로 (audit용)
}

// GetItem 단건 조회. 트랜잭션 내부에서 raw SQL DAO를 사용한다.
func GetItem(ctx context.Context, db *gorm.DB, meta Metadata, id int64) (*Item, error) {
	var item *Item
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sqlTx, err := com.SQLTx(tx)
		if err != nil {
			return err
		}
		it, err := SelectItemByID(ctx, sqlTx, meta.MemberID, id)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		item = it
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ListItems 회원의 모든 아이템.
func ListItems(ctx context.Context, db *gorm.DB, meta Metadata) ([]Item, error) {
	var items []Item
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sqlTx, err := com.SQLTx(tx)
		if err != nil {
			return err
		}
		items, err = SelectItemsByMember(ctx, sqlTx, meta.MemberID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// CreateItem 신규 INSERT 후 부여된 id 반환.
// 실제로는 트랜잭션이 필요하지 않은 케이스지만, 같은 패턴 유지를 위해 트랜잭션 wrapper 사용.
func CreateItem(ctx context.Context, db *gorm.DB, meta Metadata, name string) (int64, error) {
	var newID int64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sqlTx, err := com.SQLTx(tx)
		if err != nil {
			return err
		}
		id, err := InsertItem(ctx, sqlTx, meta.MemberID, name, meta.RegPgm)
		if err != nil {
			return err
		}
		newID = id
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newID, nil
}

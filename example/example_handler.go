package example

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-gin-lambda-starter/com"
)

// metaFrom 인증 미들웨어가 세팅한 컨텍스트 값에서 공통 메타 추출.
func metaFrom(c *gin.Context) Metadata {
	return Metadata{
		MemberID: c.GetString(com.CtxKeyMemberID),
		RegPgm:   c.FullPath(),
	}
}

// GetItemHandler GET /example/items/:id
func GetItemHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			com.BadRequest(c, "id는 정수여야 합니다")
			return
		}

		item, err := GetItem(c.Request.Context(), db, metaFrom(c), id)
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "아이템을 찾을 수 없습니다"})
			return
		}
		if err != nil {
			com.RespondServiceError(c, err)
			return
		}
		com.OK(c, gin.H{"item": item})
	}
}

// ListItemsHandler GET /example/items
func ListItemsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := ListItems(c.Request.Context(), db, metaFrom(c))
		if err != nil {
			com.RespondServiceError(c, err)
			return
		}
		// nil → 빈 배열로 정규화
		if items == nil {
			items = []Item{}
		}
		c.Set(com.CtxKeyRecordCount, len(items))
		com.OK(c, gin.H{"items": items, "count": len(items)})
	}
}

// CreateItemHandler POST /example/items
type createItemRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

func CreateItemHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			com.BadRequest(c, "name 은 필수이며 100자 이하여야 합니다")
			return
		}

		id, err := CreateItem(c.Request.Context(), db, metaFrom(c), req.Name)
		if err != nil {
			com.RespondServiceError(c, err)
			return
		}
		c.Set(com.CtxKeyRecordCount, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true, "id": id})
	}
}

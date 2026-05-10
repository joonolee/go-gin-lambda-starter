package com

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	CtxKeyMemberID    = "memberID"
	CtxKeyRecordCount = "recordCount"

	HealthCheckPath = "/health"
)

// jwks 백그라운드 고루틴이 1시간마다 Firebase 공개키 세트를 자동 갱신
var jwks *keyfunc.JWKS

// InitFirebaseJWKS Firebase 공개키 세트 초기화. main()에서 한 번 호출.
// 다른 인증 스킴으로 교체 시 이 함수와 FirebaseAuth만 바꾸면 된다.
func InitFirebaseJWKS() {
	jwksURL := "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"

	options := keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  10 * time.Second,
	}

	var err error
	jwks, err = keyfunc.Get(jwksURL, options)
	if err != nil {
		log.Fatalf("JWKS 초기화 실패: %v", err)
	}
}

func CloseJWKS() {
	if jwks != nil {
		jwks.EndBackground()
	}
}

// FirebaseAuth Firebase ID Token 검증 미들웨어.
// 다른 JWT 발급자로 교체하려면 expectedIssuer/audience와 jwksURL만 변경.
func FirebaseAuth(firebaseProjectID string) gin.HandlerFunc {
	expectedIssuer := fmt.Sprintf("https://securetoken.google.com/%s", firebaseProjectID)
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Authorization 헤더가 필요합니다"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Bearer 토큰 형식이 필요합니다"})
			return
		}
		tokenString := parts[1]

		if jwks == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "인증 서비스 초기화 실패"})
			return
		}

		token, err := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithIssuer(expectedIssuer),
			jwt.WithAudience(firebaseProjectID),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "유효하지 않은 토큰입니다"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "토큰 클레임 파싱 실패"})
			return
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "사용자 ID가 없습니다"})
			return
		}

		c.Set(CtxKeyMemberID, sub)
		c.Next()
	}
}

// Recovery panic을 500 응답으로 변환.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("패닉 발생: %v", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "서버 내부 오류"})
	})
}

// RespondServiceError 서비스 오류 응답 공통 처리.
// 내부 에러는 로그에만 남기고 클라이언트에는 일반 메시지만 노출한다.
func RespondServiceError(c *gin.Context, err error) {
	log.Printf("서비스 오류 [%s]: %v", c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "서버 내부 오류"})
}

// ApiLogging 모든 요청을 api_log 테이블에 기록. 헬스 체크는 제외.
func ApiLogging(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if c.Request.URL.Path == HealthCheckPath {
			return
		}

		latencyMs := int(time.Since(start).Milliseconds())
		entry := ApiLog{
			Method:     c.Request.Method,
			Endpoint:   c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			LatencyMs:  latencyMs,
		}

		if memberID := c.GetString(CtxKeyMemberID); memberID != "" {
			entry.MemberID = &memberID
		}
		if count, exists := c.Get(CtxKeyRecordCount); exists {
			if n, ok := count.(int); ok {
				entry.RecordCount = &n
			}
		}

		if err := db.Create(&entry).Error; err != nil {
			log.Printf("API 로그 저장 실패: %v", err)
		}
	}
}

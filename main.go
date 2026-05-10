package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	"go-gin-lambda-starter/com"
	"go-gin-lambda-starter/example"
)

func main() {
	// --- 설정 초기화 ---
	cfg := com.LoadConfig()
	db := com.InitDB(cfg)

	// --- 인증 미들웨어 ---
	gin.SetMode(gin.ReleaseMode)
	if cfg.FirebaseProjectID == "" {
		log.Fatal("FIREBASE_PROJECT_ID 환경변수가 설정되지 않았습니다")
	}
	com.InitFirebaseJWKS()
	defer com.CloseJWKS()
	authMiddleware := com.FirebaseAuth(cfg.FirebaseProjectID)

	// --- 라우터 설정 ---
	router := gin.New()
	router.Use(com.Recovery())
	router.Use(gin.Logger())
	router.Use(com.ApiLogging(db))

	// --- 공개 라우트 ---
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// --- 인증 라우트 ---
	authorized := router.Group("")
	authorized.Use(authMiddleware)

	// --- example 모듈 (샘플 3-tier 패턴) ---
	exampleGroup := authorized.Group("/example")
	exampleGroup.GET("/items", example.ListItemsHandler(db))
	exampleGroup.GET("/items/:id", example.GetItemHandler(db))
	exampleGroup.POST("/items", example.CreateItemHandler(db))

	// --- 서버 시작 ---
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		log.Println("Lambda 모드로 시작")
		adapter := ginadapter.NewV2(router)
		lambda.Start(adapter.ProxyWithContext)
	} else {
		log.Println("로컬 서버 시작: http://localhost:8080")
		log.Fatal(router.Run(":8080"))
	}
}

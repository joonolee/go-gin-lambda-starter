package com

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 표준 응답 envelope: {"ok": true/false, ...}
// handler에서 직접 c.JSON(http.StatusOK, gin.H{"ok": true, ...}) 써도 무방.
// helper는 자주 쓰는 형태만 짧게 모아둔 것.

// OK 200 OK + {"ok": true}. extra가 있으면 같은 객체에 병합.
func OK(c *gin.Context, extra gin.H) {
	body := gin.H{"ok": true}
	for k, v := range extra {
		body[k] = v
	}
	c.JSON(http.StatusOK, body)
}

// OKWithCount 동기화/대량 처리 결과처럼 카운트만 돌려줄 때.
func OKWithCount(c *gin.Context, count int) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "count": count})
}

// BadRequest 400 + {"ok": false, "error": msg}.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
}

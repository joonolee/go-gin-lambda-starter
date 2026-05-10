package com_test

import (
	"strings"
	"testing"

	"go-gin-lambda-starter/com"
)

func TestStaticQuery(t *testing.T) {
	sql := `
-- name: static
/* static : 정적 쿼리 */
select now()
`
	qs := com.ParseNamedQueries(sql)
	sqlText, args, err := qs.MustGet("static").Render(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sqlText, "select now()") {
		t.Fatalf("expected sql to contain 'select now()', got %q", sqlText)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestNamedParams(t *testing.T) {
	sql := `
-- name: insert_rec
/* insert_rec : 레코드 삽입 */
insert into t_foo (member_id, val) values (:member_id, :val)
`
	qs := com.ParseNamedQueries(sql)
	sqlText, args, err := qs.MustGet("insert_rec").Render(map[string]any{
		"member_id": "user1",
		"val":       42,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sqlText, "($1, $2)") {
		t.Fatalf("unexpected sql: %q", sqlText)
	}
	if len(args) != 2 || args[0] != "user1" || args[1] != 42 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSameParamReused(t *testing.T) {
	sql := `
-- name: reuse
/* reuse : 같은 파라미터 두 번 사용 */
select :x + :x
`
	qs := com.ParseNamedQueries(sql)
	sqlText, args, err := qs.MustGet("reuse").Render(map[string]any{
		"x": 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sqlText, "$1 + $1") {
		t.Fatalf("unexpected sql: %q", sqlText)
	}
	if len(args) != 1 || args[0] != 10 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestMissingParam(t *testing.T) {
	sql := `
-- name: missing
/* missing : 누락 파라미터 */
select :a, :b
`
	qs := com.ParseNamedQueries(sql)
	_, _, err := qs.MustGet("missing").Render(map[string]any{
		"a": 1,
	})
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestIfBlockIncluded(t *testing.T) {
	sql := `
-- name: conditional
/* conditional : 조건 블록 포함 */
select id from t_foo
where member_id = :member_id
{{if .ids}}
	and id = any(:ids)
{{end}}
`
	qs := com.ParseNamedQueries(sql)

	sqlText, args, err := qs.MustGet("conditional").Render(map[string]any{
		"member_id": "u1",
		"ids":       []int64{1, 2, 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	_ = sqlText
}

func TestIfBlockExcluded(t *testing.T) {
	sql := `
-- name: conditional
/* conditional : 조건 블록 제외 */
select id from t_foo
where member_id = :member_id
{{if .ids}}
	and id = any(:ids)
{{end}}
`
	qs := com.ParseNamedQueries(sql)

	sqlText, args, err := qs.MustGet("conditional").Render(map[string]any{
		"member_id": "u1",
		"ids":       nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg (only member_id), got %d: %v", len(args), args)
	}
	_ = sqlText
}

func TestMustGetPanicsOnMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing query key")
		}
	}()
	qs := com.ParseNamedQueries("-- name: exists\nselect 1\n")
	qs.MustGet("does_not_exist")
}

func TestTemplateParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid template")
		}
	}()
	com.ParseNamedQueries("-- name: bad\n{{if .x\nselect 1\n")
}

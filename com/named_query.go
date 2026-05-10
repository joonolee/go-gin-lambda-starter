package com

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

var (
	namedQueryMarker = regexp.MustCompile(`(?m)^\s*--\s*name:\s*(\S+)\s*$`)
	namedParamRE     = regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
)

type Query struct {
	name string
	tpl  *template.Template
}

type Queries map[string]*Query

// MustGet 누락 키는 panic — DAO init에서 호출하면 빌드 후 즉시 발견.
func (qs Queries) MustGet(name string) *Query {
	q, ok := qs[name]
	if !ok || q == nil {
		panic("쿼리 누락: " + name)
	}
	return q
}

// ParseNamedQueries `-- name: <key>` 마커로 구분된 SQL 블록을 파싱한다.
// 본문은 text/template으로 처리되어 {{if}} {{range}} 등 조건 블록을 지원한다.
// template 파싱 에러는 init 시점 panic.
func ParseNamedQueries(content string) Queries {
	out := Queries{}
	var name string
	var buf strings.Builder
	flush := func() {
		if name == "" {
			return
		}
		body := strings.TrimSpace(buf.String())
		tpl, err := template.New(name).Parse(body)
		if err != nil {
			panic(fmt.Sprintf("쿼리 template 파싱 실패 [%s]: %v", name, err))
		}
		out[name] = &Query{name: name, tpl: tpl}
	}
	for _, line := range strings.Split(content, "\n") {
		if m := namedQueryMarker.FindStringSubmatch(line); m != nil {
			flush()
			name = m[1]
			buf.Reset()
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	flush()
	return out
}

// Render template 평가 후 :param → $N 치환. 인자 슬라이스를 함께 반환한다.
// 같은 :param 이 여러 번 등장하면 같은 $N 으로 매핑되고 args에는 한 번만 추가된다.
func (q *Query) Render(params map[string]any) (string, []any, error) {
	var buf bytes.Buffer
	if err := q.tpl.Execute(&buf, params); err != nil {
		return "", nil, fmt.Errorf("쿼리 렌더 실패 [%s]: %w", q.name, err)
	}
	// "::" PostgreSQL 타입 캐스트가 named param 정규식에 잘못 매칭되는 것을 방지
	const dblColon = "\x00::"
	rendered := strings.ReplaceAll(buf.String(), "::", dblColon)
	var args []any
	seen := map[string]int{}
	var renderErr error
	out := namedParamRE.ReplaceAllStringFunc(rendered, func(m string) string {
		key := m[1:]
		if idx, ok := seen[key]; ok {
			return fmt.Sprintf("$%d", idx)
		}
		v, ok := params[key]
		if !ok {
			if renderErr == nil {
				renderErr = fmt.Errorf("쿼리 [%s]: 파라미터 누락 :%s", q.name, key)
			}
			return m
		}
		args = append(args, v)
		idx := len(args)
		seen[key] = idx
		return fmt.Sprintf("$%d", idx)
	})
	if renderErr != nil {
		return "", nil, renderErr
	}
	return strings.ReplaceAll(out, dblColon, "::"), args, nil
}

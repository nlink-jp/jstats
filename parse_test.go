package main

import (
	"testing"
)

func TestParseExpr_Count(t *testing.T) {
	q, err := parseExpr("count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Funcs) != 1 || q.Funcs[0].Func != "count" || q.Funcs[0].Field != "" {
		t.Errorf("unexpected funcs: %+v", q.Funcs)
	}
	if len(q.ByFields) != 0 {
		t.Errorf("unexpected by fields: %v", q.ByFields)
	}
}

func TestParseExpr_CountByField(t *testing.T) {
	q, err := parseExpr("count by status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Funcs) != 1 || q.Funcs[0].Func != "count" {
		t.Errorf("unexpected funcs: %+v", q.Funcs)
	}
	if len(q.ByFields) != 1 || q.ByFields[0] != "status" {
		t.Errorf("unexpected by fields: %v", q.ByFields)
	}
}

func TestParseExpr_MultipleAggs(t *testing.T) {
	q, err := parseExpr("count, avg(latency), max(latency) by host, service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Funcs) != 3 {
		t.Fatalf("expected 3 funcs, got %d", len(q.Funcs))
	}
	if q.Funcs[0].Func != "count" {
		t.Errorf("func[0]: want count, got %s", q.Funcs[0].Func)
	}
	if q.Funcs[1].Func != "avg" || q.Funcs[1].Field != "latency" {
		t.Errorf("func[1]: %+v", q.Funcs[1])
	}
	if q.Funcs[2].Func != "max" || q.Funcs[2].Field != "latency" {
		t.Errorf("func[2]: %+v", q.Funcs[2])
	}
	if len(q.ByFields) != 2 || q.ByFields[0] != "host" || q.ByFields[1] != "service" {
		t.Errorf("by fields: %v", q.ByFields)
	}
}

func TestParseExpr_Alias(t *testing.T) {
	q, err := parseExpr("avg(latency) as avg_ms, count as total")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Funcs[0].Alias != "avg_ms" {
		t.Errorf("alias[0]: want avg_ms, got %q", q.Funcs[0].Alias)
	}
	if q.Funcs[1].Alias != "total" {
		t.Errorf("alias[1]: want total, got %q", q.Funcs[1].Alias)
	}
}

func TestParseExpr_Percentile(t *testing.T) {
	q, err := parseExpr("p95(latency), p99(latency)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Funcs[0].Func != "p" || q.Funcs[0].Perc != 95 || q.Funcs[0].Field != "latency" {
		t.Errorf("p95: %+v", q.Funcs[0])
	}
	if q.Funcs[1].Func != "p" || q.Funcs[1].Perc != 99 {
		t.Errorf("p99: %+v", q.Funcs[1])
	}
}

func TestParseExpr_OutputName(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"count", "count"},
		{"count(field)", "count(field)"},
		{"avg(x)", "avg(x)"},
		{"p95(latency)", "p95(latency)"},
		{"avg(x) as mean", "mean"},
	}
	for _, c := range cases {
		q, err := parseExpr(c.expr)
		if err != nil {
			t.Fatalf("%q: %v", c.expr, err)
		}
		got := q.Funcs[0].OutputName()
		if got != c.want {
			t.Errorf("%q: OutputName() = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestParseExpr_CaseInsensitive(t *testing.T) {
	q, err := parseExpr("COUNT, AVG(latency) AS Avg BY host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Funcs[0].Func != "count" {
		t.Errorf("func[0].Func: want count, got %s", q.Funcs[0].Func)
	}
	if q.Funcs[1].Func != "avg" {
		t.Errorf("func[1].Func: want avg, got %s", q.Funcs[1].Func)
	}
	if q.Funcs[1].Alias != "Avg" {
		t.Errorf("alias: want Avg, got %q", q.Funcs[1].Alias)
	}
	if q.ByFields[0] != "host" {
		t.Errorf("by[0]: want host, got %q", q.ByFields[0])
	}
}

func TestParseExpr_Errors(t *testing.T) {
	cases := []string{
		"unknown_func(x)",
		"avg()",    // empty field
		"p999(x)", // percentile out of range
	}
	for _, expr := range cases {
		if _, err := parseExpr(expr); err == nil {
			t.Errorf("%q: expected error, got nil", expr)
		}
	}
}

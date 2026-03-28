package main

import (
	"math"
	"testing"
)

// ---- helpers ---------------------------------------------------------------

func rows(fields ...map[string]interface{}) []Row {
	out := make([]Row, len(fields))
	for i, f := range fields {
		out[i] = f
	}
	return out
}

func r(kv ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func approxEq(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// ---- count -----------------------------------------------------------------

func TestCount_All(t *testing.T) {
	fn := AggFunc{Func: "count"}
	val, err := computeAgg(fn, rows(r("x", 1.0), r("x", 2.0), r("x", 3.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 3 {
		t.Errorf("want 3, got %v", val)
	}
}

func TestCount_Field(t *testing.T) {
	fn := AggFunc{Func: "count", Field: "x"}
	// One row has nil x
	val, err := computeAgg(fn, rows(r("x", 1.0), r("y", 2.0), r("x", 3.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 2 {
		t.Errorf("want 2, got %v", val)
	}
}

// ---- sum / min / max / avg -------------------------------------------------

func TestSum(t *testing.T) {
	fn := AggFunc{Func: "sum", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 1.0), r("v", 2.0), r("v", 3.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 6 {
		t.Errorf("want 6, got %v", val)
	}
}

func TestMin(t *testing.T) {
	fn := AggFunc{Func: "min", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 5.0), r("v", 2.0), r("v", 8.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 2 {
		t.Errorf("want 2, got %v", val)
	}
}

func TestMax(t *testing.T) {
	fn := AggFunc{Func: "max", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 5.0), r("v", 2.0), r("v", 8.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 8 {
		t.Errorf("want 8, got %v", val)
	}
}

func TestAvg(t *testing.T) {
	fn := AggFunc{Func: "avg", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 1.0), r("v", 2.0), r("v", 3.0)))
	if err != nil {
		t.Fatal(err)
	}
	if !approxEq(val.(float64), 2.0, 1e-9) {
		t.Errorf("want 2, got %v", val)
	}
}

// ---- median / percentile ---------------------------------------------------

func TestMedian_Odd(t *testing.T) {
	fn := AggFunc{Func: "median", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 3.0), r("v", 1.0), r("v", 2.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 2 {
		t.Errorf("want 2, got %v", val)
	}
}

func TestMedian_Even(t *testing.T) {
	fn := AggFunc{Func: "median", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 1.0), r("v", 2.0), r("v", 3.0), r("v", 4.0)))
	if err != nil {
		t.Fatal(err)
	}
	if !approxEq(val.(float64), 2.5, 1e-9) {
		t.Errorf("want 2.5, got %v", val)
	}
}

func TestPercentile_P0_P100(t *testing.T) {
	data := rows(r("v", 1.0), r("v", 2.0), r("v", 3.0), r("v", 4.0), r("v", 5.0))

	p0, _ := computeAgg(AggFunc{Func: "p", Field: "v", Perc: 0}, data)
	if p0.(float64) != 1 {
		t.Errorf("p0: want 1, got %v", p0)
	}

	p100, _ := computeAgg(AggFunc{Func: "p", Field: "v", Perc: 100}, data)
	if p100.(float64) != 5 {
		t.Errorf("p100: want 5, got %v", p100)
	}
}

// ---- stdev / var -----------------------------------------------------------

func TestStdev(t *testing.T) {
	// Sample stdev of [2, 4, 4, 4, 5, 5, 7, 9]:
	// mean=5, sum_sq_dev=32, sample variance=32/7≈4.571, sample stdev≈2.138
	fn := AggFunc{Func: "stdev", Field: "v"}
	data := rows(r("v", 2.0), r("v", 4.0), r("v", 4.0), r("v", 4.0),
		r("v", 5.0), r("v", 5.0), r("v", 7.0), r("v", 9.0))
	val, err := computeAgg(fn, data)
	if err != nil {
		t.Fatal(err)
	}
	if !approxEq(val.(float64), 2.1381, 0.001) {
		t.Errorf("want ~2.138, got %v", val)
	}
}

func TestVar(t *testing.T) {
	fn := AggFunc{Func: "var", Field: "v"}
	data := rows(r("v", 2.0), r("v", 4.0), r("v", 4.0), r("v", 4.0),
		r("v", 5.0), r("v", 5.0), r("v", 7.0), r("v", 9.0))
	val, err := computeAgg(fn, data)
	if err != nil {
		t.Fatal(err)
	}
	// sample variance = 32/7 ≈ 4.571
	if !approxEq(val.(float64), 4.5714, 0.001) {
		t.Errorf("want ~4.571, got %v", val)
	}
}

// ---- range -----------------------------------------------------------------

func TestRange(t *testing.T) {
	fn := AggFunc{Func: "range", Field: "v"}
	val, err := computeAgg(fn, rows(r("v", 3.0), r("v", 10.0), r("v", 1.0)))
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 9 {
		t.Errorf("want 9, got %v", val)
	}
}

// ---- dc / first / last / mode ----------------------------------------------

func TestDC(t *testing.T) {
	fn := AggFunc{Func: "dc", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "a"), r("s", "b"), r("s", "a")))
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 2 {
		t.Errorf("want 2, got %v", val)
	}
}

func TestFirst(t *testing.T) {
	fn := AggFunc{Func: "first", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "x"), r("s", "y"), r("s", "z")))
	if err != nil {
		t.Fatal(err)
	}
	if val.(string) != "x" {
		t.Errorf("want x, got %v", val)
	}
}

func TestLast(t *testing.T) {
	fn := AggFunc{Func: "last", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "x"), r("s", "y"), r("s", "z")))
	if err != nil {
		t.Fatal(err)
	}
	if val.(string) != "z" {
		t.Errorf("want z, got %v", val)
	}
}

func TestMode(t *testing.T) {
	fn := AggFunc{Func: "mode", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "a"), r("s", "b"), r("s", "a"), r("s", "c"), r("s", "a")))
	if err != nil {
		t.Fatal(err)
	}
	if val.(string) != "a" {
		t.Errorf("want a, got %v", val)
	}
}

// ---- values / list ---------------------------------------------------------

func TestValues(t *testing.T) {
	fn := AggFunc{Func: "values", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "a"), r("s", "b"), r("s", "a")))
	if err != nil {
		t.Fatal(err)
	}
	got := val.([]interface{})
	if len(got) != 2 {
		t.Errorf("want 2 distinct values, got %v", got)
	}
}

func TestList(t *testing.T) {
	fn := AggFunc{Func: "list", Field: "s"}
	val, err := computeAgg(fn, rows(r("s", "a"), r("s", "b"), r("s", "a")))
	if err != nil {
		t.Fatal(err)
	}
	got := val.([]interface{})
	if len(got) != 3 {
		t.Errorf("want 3 values, got %v", got)
	}
}

// ---- grouping --------------------------------------------------------------

func TestGroupBy(t *testing.T) {
	q, _ := parseExpr("count by status")
	data := rows(
		r("status", "200"),
		r("status", "500"),
		r("status", "200"),
		r("status", "200"),
	)
	result, headers, err := computeStats(data, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("want 2 groups, got %d", len(result))
	}
	if headers[0] != "status" || headers[1] != "count" {
		t.Errorf("headers: %v", headers)
	}
	// First group (200) should have count 3
	if result[0]["count"].(int) != 3 {
		t.Errorf("200 count: want 3, got %v", result[0]["count"])
	}
}

func TestNoGroupBy(t *testing.T) {
	q, _ := parseExpr("count, sum(v)")
	data := rows(r("v", 1.0), r("v", 2.0), r("v", 3.0))
	result, _, err := computeStats(data, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("want 1 group, got %d", len(result))
	}
	if result[0]["count"].(int) != 3 {
		t.Errorf("count: want 3, got %v", result[0]["count"])
	}
	if result[0]["sum(v)"].(float64) != 6 {
		t.Errorf("sum: want 6, got %v", result[0]["sum(v)"])
	}
}

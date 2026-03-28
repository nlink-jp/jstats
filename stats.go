package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// computeStats groups rows and applies aggregation functions.
// Returns result rows, column headers (in declaration order), and any error.
func computeStats(rows []Row, q StatsQuery) ([]Row, []string, error) {
	// Build column header list: by fields first, then agg output names.
	headers := make([]string, 0, len(q.ByFields)+len(q.Funcs))
	headers = append(headers, q.ByFields...)
	for _, fn := range q.Funcs {
		headers = append(headers, fn.OutputName())
	}

	// Group rows.
	type group struct {
		key    string
		sample Row   // first row of group (for by-field values)
		rows   []Row
	}
	var groupOrder []string
	groups := map[string]*group{}

	for _, row := range rows {
		key := groupKey(row, q.ByFields)
		if _, ok := groups[key]; !ok {
			groups[key] = &group{key: key, sample: row}
			groupOrder = append(groupOrder, key)
		}
		groups[key].rows = append(groups[key].rows, row)
	}

	result := make([]Row, 0, len(groupOrder))
	for _, key := range groupOrder {
		g := groups[key]
		out := make(Row, len(headers))

		// Copy by-field values from first row of group.
		for _, f := range q.ByFields {
			out[f] = g.sample[f]
		}

		// Compute each aggregate.
		for _, fn := range q.Funcs {
			val, err := computeAgg(fn, g.rows)
			if err != nil {
				return nil, nil, fmt.Errorf("function %s: %w", fn.OutputName(), err)
			}
			out[fn.OutputName()] = val
		}

		result = append(result, out)
	}

	return result, headers, nil
}

// groupKey builds a stable string key from the group-by fields of a row.
func groupKey(row Row, fields []string) string {
	if len(fields) == 0 {
		return "__all__"
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = strVal(row[f])
	}
	return strings.Join(parts, "\x00")
}

// computeAgg computes one aggregation function over a slice of rows.
func computeAgg(fn AggFunc, rows []Row) (interface{}, error) {
	switch fn.Func {
	case "count":
		if fn.Field == "" {
			return len(rows), nil
		}
		n := 0
		for _, r := range rows {
			if v, ok := r[fn.Field]; ok && v != nil {
				n++
			}
		}
		return n, nil

	case "sum":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		s := 0.0
		for _, v := range vals {
			s += v
		}
		return roundFloat(s), nil

	case "min":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return roundFloat(m), nil

	case "max":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		m := vals[0]
		for _, v := range vals[1:] {
			if v > m {
				m = v
			}
		}
		return roundFloat(m), nil

	case "avg":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		return roundFloat(mean(vals)), nil

	case "median":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		return roundFloat(percentile(vals, 50)), nil

	case "p":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		return roundFloat(percentile(vals, fn.Perc)), nil

	case "stdev":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) < 2 {
			return nil, nil
		}
		return roundFloat(stddev(vals, false)), nil

	case "var":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) < 2 {
			return nil, nil
		}
		return roundFloat(variance(vals, false)), nil

	case "range":
		vals, err := floatVals(rows, fn.Field)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			return nil, nil
		}
		lo, hi := vals[0], vals[0]
		for _, v := range vals[1:] {
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
		return roundFloat(hi - lo), nil

	case "dc":
		seen := map[string]struct{}{}
		for _, r := range rows {
			v, ok := r[fn.Field]
			if !ok || v == nil {
				continue
			}
			seen[strVal(v)] = struct{}{}
		}
		return len(seen), nil

	case "first":
		for _, r := range rows {
			if v, ok := r[fn.Field]; ok {
				return v, nil
			}
		}
		return nil, nil

	case "last":
		for i := len(rows) - 1; i >= 0; i-- {
			if v, ok := rows[i][fn.Field]; ok {
				return v, nil
			}
		}
		return nil, nil

	case "mode":
		return modeVal(rows, fn.Field), nil

	case "values":
		return distinctValues(rows, fn.Field), nil

	case "list":
		return allValues(rows, fn.Field), nil
	}

	return nil, fmt.Errorf("unsupported function %q", fn.Func)
}

// ---- helpers ---------------------------------------------------------------

func floatVals(rows []Row, field string) ([]float64, error) {
	out := make([]float64, 0, len(rows))
	for _, r := range rows {
		v, ok := r[field]
		if !ok || v == nil {
			continue
		}
		f, err := toFloat(v)
		if err != nil {
			return nil, fmt.Errorf("field %q value %v: %w", field, v, err)
		}
		out = append(out, f)
	}
	return out, nil
}

func toFloat(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case json_number:
		return n.Float64()
	}
	return 0, fmt.Errorf("not a number (type %T)", v)
}

// json_number is encoding/json's Number type (used when decoding with UseNumber).
type json_number interface {
	Float64() (float64, error)
}

func mean(vals []float64) float64 {
	s := 0.0
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

func variance(vals []float64, population bool) float64 {
	m := mean(vals)
	s := 0.0
	for _, v := range vals {
		d := v - m
		s += d * d
	}
	n := float64(len(vals))
	if !population {
		n--
	}
	return s / n
}

func stddev(vals []float64, population bool) float64 {
	return math.Sqrt(variance(vals, population))
}

func percentile(vals []float64, p int) float64 {
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	// Linear interpolation
	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func modeVal(rows []Row, field string) interface{} {
	freq := map[string]int{}
	var order []string
	vals := map[string]interface{}{}

	for _, r := range rows {
		v, ok := r[field]
		if !ok || v == nil {
			continue
		}
		k := strVal(v)
		if _, seen := freq[k]; !seen {
			order = append(order, k)
			vals[k] = v
		}
		freq[k]++
	}

	maxFreq := 0
	var result interface{}
	for _, k := range order {
		if freq[k] > maxFreq {
			maxFreq = freq[k]
			result = vals[k]
		}
	}
	return result
}

func distinctValues(rows []Row, field string) []interface{} {
	seen := map[string]bool{}
	var out []interface{}
	for _, r := range rows {
		v, ok := r[field]
		if !ok || v == nil {
			continue
		}
		k := strVal(v)
		if !seen[k] {
			seen[k] = true
			out = append(out, v)
		}
	}
	return out
}

func allValues(rows []Row, field string) []interface{} {
	var out []interface{}
	for _, r := range rows {
		v, ok := r[field]
		if !ok || v == nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

func roundFloat(f float64) float64 {
	// Round to 6 significant decimal digits to avoid floating-point noise.
	if f == 0 {
		return 0
	}
	pow := math.Pow(10, 6-math.Floor(math.Log10(math.Abs(f)))-1)
	return math.Round(f*pow) / pow
}

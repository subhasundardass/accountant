// Package expr implements the GoDL expression language: the right-hand side
// of attributes such as "Set as : if $$IsEmpty:$Parent then "Primary" else $Parent".
//
// Grammar (loosest to tightest binding):
//
//	or
//	and
//	=  <>  !=  <  <=  >  >=
//	+  -
//	*  /
//	not x   -x   +x   $$Func:arg:arg   primary
//
// Primaries:
//
//	123  12.5  "text"  ""escaped""  Yes  No  true  false
//	$Name           method of the current object
//	##Name          variable
//	@Name           named formula, evaluated in the current context
//	$$Func:a:b      built-in function; arguments are unary-level terms,
//	                so use parentheses for anything bigger: $$String:($A + $B)
//	( expr )
//	if c then a else b [endif]
//
// Values are nil, string, float64, bool or time.Time.
//
// Money warning: numbers are float64 here. That is fine for layout logic and
// comparisons (== uses a 1e-9 tolerance), but accounting totals should come
// from the data layer as exact integers/decimals, not be summed in expressions.
package expr

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Value is the dynamic type of every expression result.
type Value = any

// Normalize converts the numeric types Go code tends to hand us (int, int64,
// float32, ...) to float64 so the evaluator only deals with one number type.
func Normalize(v Value) Value {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	case float32:
		return float64(n)
	}
	return v
}

// ToNumber converts v to a number. nil is 0. Strings are parsed. ok is false
// for values with no numeric meaning.
func ToNumber(v Value) (float64, bool) {
	switch n := Normalize(v).(type) {
	case nil:
		return 0, true
	case float64:
		return n, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case bool:
		if n {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// ToString renders a value the way a report shows it.
func ToString(v Value) string {
	switch s := Normalize(v).(type) {
	case nil:
		return ""
	case string:
		return s
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case bool:
		if s {
			return "Yes"
		}
		return "No"
	case time.Time:
		return s.Format("02-Jan-2006")
	}
	return fmt.Sprint(v)
}

// ToBool is the truthiness rule used by if/and/or/not.
func ToBool(v Value) bool {
	switch b := Normalize(v).(type) {
	case nil:
		return false
	case bool:
		return b
	case float64:
		return b != 0
	case string:
		switch strings.ToLower(strings.TrimSpace(b)) {
		case "", "no", "false", "0":
			return false
		}
		return true
	}
	return true
}

// IsEmpty: nil, a blank string, or zero.
func IsEmpty(v Value) bool {
	switch s := Normalize(v).(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(s) == ""
	case float64:
		return s == 0
	case time.Time:
		return s.IsZero()
	}
	return false
}

// compare orders two values: -1, 0 or 1. Rules, in order: two times compare
// as times; if either side is a bool, as bools; if either side is a number and
// both convert, as numbers (equal within 1e-9); otherwise as strings,
// case-insensitively (TDL string comparison ignores case). nil acts as the
// zero value of the other side.
func compare(a, b Value) int {
	a, b = Normalize(a), Normalize(b)
	if t1, ok := a.(time.Time); ok {
		if t2, ok := b.(time.Time); ok {
			switch {
			case t1.Before(t2):
				return -1
			case t1.After(t2):
				return 1
			}
			return 0
		}
	}
	_, aBool := a.(bool)
	_, bBool := b.(bool)
	if aBool || bBool {
		x, y := ToBool(a), ToBool(b)
		switch {
		case x == y:
			return 0
		case !x:
			return -1
		}
		return 1
	}
	_, aNum := a.(float64)
	_, bNum := b.(float64)
	if aNum || bNum {
		x, okx := ToNumber(a)
		y, oky := ToNumber(b)
		if okx && oky {
			switch {
			case math.Abs(x-y) < 1e-9:
				return 0
			case x < y:
				return -1
			}
			return 1
		}
	}
	return strings.Compare(strings.ToLower(ToString(a)), strings.ToLower(ToString(b)))
}

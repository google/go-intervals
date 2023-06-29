package modinterval

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var cmpOpts = []cmp.Option{
	cmpopts.EquateEmpty(),
	cmp.Transformer("RealIntervalString", func(i RealIntInterval) string {
		return i.String()
	}),
}

func ExampleIntInterval_basics() {
	fmt.Printf("%s\n", FromStartSizeInt(Modulus(10), 9, 4))
	fmt.Printf("interval.Contains(2) = %v\n", FromStartSizeInt(Modulus(10), 9, 4).Contains(2))
	fmt.Printf("interval.Contains(9) = %v\n", FromStartSizeInt(Modulus(10), 9, 4).Contains(9))
	fmt.Printf("interval.Contains(3) = %v\n", FromStartSizeInt(Modulus(10), 9, 4).Contains(3))
	// Output:
	// <mod=10; [9, 9], [0, 2]>
	// interval.Contains(2) = true
	// interval.Contains(9) = true
	// interval.Contains(3) = false
}

func ExampleIntInterval_RealIntervals() {
	modularInterval := FromStartSizeInt(Modulus(6), 5, 4)
	realIntervals := modularInterval.RealIntervals()
	fmt.Printf("%s\n", modularInterval)
	fmt.Printf(".RealIntervals()[0] = %s\n", realIntervals[0])
	fmt.Printf(".RealIntervals()[1] = %s\n", realIntervals[1])
	// Output:
	// <mod=6; [5, 5], [0, 2]>
	// .RealIntervals()[0] = [5, 5]
	// .RealIntervals()[1] = [0, 2]
}

func ExampleIntInterval_RealIntervals_empty() {
	modularInterval := FromStartSizeInt(Modulus(100), 5, 0)
	fmt.Printf("{empty}.RealIntervals() has length %d", len(modularInterval.RealIntervals()))
	// Output:
	// {empty}.RealIntervals() has length 0
}

func ExampleIntInterval_RealIntervals_nonwrapping() {
	m := FromStartSizeInt(Modulus(100), 5, 20)
	r := m.RealIntervals()
	fmt.Printf(".RealIntervals() has length %d: %s", len(r), r[0])
	// Output:
	// .RealIntervals() has length 1: [5, 24]
}

func ExampleModulus_ArrayOffset() {
	for _, tt := range []struct {
		m   Modulus
		arg int
	}{
		{10, -1},
		{10, -11},
		{10, 1},
		{10, 10},
		{10, 11},
	} {
		fmt.Printf("Modulus(%d).ArrayOffset(%d) = %d\n", tt.m, tt.arg, tt.m.ArrayOffset(tt.arg))
	}
	// Output:
	// Modulus(10).ArrayOffset(-1) = 9
	// Modulus(10).ArrayOffset(-11) = 9
	// Modulus(10).ArrayOffset(1) = 1
	// Modulus(10).ArrayOffset(10) = 0
	// Modulus(10).ArrayOffset(11) = 1
}

func ExampleModulus_IntervalSizeForward() {
	for _, tt := range []struct {
		m    Modulus
		a, b int
	}{
		{m: 10, a: 10, b: 14},
		{m: 10, a: 9, b: 7},
		{m: 10, a: 10, b: 100},
	} {
		m, a, b := tt.m, tt.a, tt.b
		fmt.Printf("Modulus(%d).IntervalSizeForward(%d, %d) = %d\n", m, a, b, m.IntervalSizeForward(a, b))
		fmt.Printf("Modulus(%d).IntervalSizeMin(%d, %d) = %d\n", m, a, b, m.IntervalSizeMin(a, b))
	}
	// Output:
	// Modulus(10).IntervalSizeForward(10, 14) = 4
	// Modulus(10).IntervalSizeMin(10, 14) = 4
	// Modulus(10).IntervalSizeForward(9, 7) = 8
	// Modulus(10).IntervalSizeMin(9, 7) = 2
	// Modulus(10).IntervalSizeForward(10, 100) = 0
	// Modulus(10).IntervalSizeMin(10, 100) = 0
}

func TestIntInterval(t *testing.T) {
	type containsCase struct {
		arg  int
		want bool
	}
	type expandStartCase struct {
		args []int
		want IntInterval
	}

	for _, tt := range []struct {
		name                  string
		iv                    IntInterval
		wantString            string
		wantIsEmpty           bool
		wantIsComplete        bool
		wantSize              int
		wantStart             int
		wantEnd               int
		wantRealIntervals     []RealIntInterval
		containsCases         []containsCase
		containsExactIntCases []containsCase
		expandStartCases      []expandStartCase
	}{
		{
			name:        "empty",
			iv:          FromStartSizeInt(10, 0, 0),
			wantString:  "<mod=10; empty>",
			wantSize:    0,
			wantIsEmpty: true,
			containsExactIntCases: []containsCase{
				{arg: 0, want: false},
				{arg: -1, want: false},
			},
			expandStartCases: []expandStartCase{
				{
					args: []int{3, 1},
					want: FromStartSizeInt(10, 1, 9),
				},
				{
					args: []int{1, 3},
					want: FromStartSizeInt(10, 1, 9),
				},
				{
					args: []int{},
					want: FromStartSizeInt(10, 0, 0),
				},
			},
		},
		{
			name:        "from 3 size 0",
			iv:          FromStartSizeInt(7, 3, 0),
			wantString:  "<mod=7; empty>",
			wantSize:    0,
			wantIsEmpty: true,
			wantStart:   3,
			wantEnd:     3,
		},
		{
			name:        "from 3 size 0 expanded",
			iv:          FromStartSizeInt(7, 3, 0).ExpandEnd(),
			wantString:  "<mod=7; empty>",
			wantSize:    0,
			wantIsEmpty: true,
			wantStart:   3,
			wantEnd:     3,
		},
		{
			name:        "from 3 size 0 expanded 2",
			iv:          FromStartSizeInt(7, 3, 0).ExpandStart(),
			wantString:  "<mod=7; empty>",
			wantSize:    0,
			wantIsEmpty: true,
			wantStart:   3,
			wantEnd:     3,
		},
		{
			name:        "from -1 size 0",
			iv:          FromStartSizeInt(7, -1, 0),
			wantString:  "<mod=7; empty>",
			wantSize:    0,
			wantIsEmpty: true,
			wantStart:   6,
			wantEnd:     6,
		},
		{
			name:        "345",
			iv:          FromStartSizeInt(10, 3, 3),
			wantString:  "<mod=10; [3, 5]>",
			wantSize:    3,
			wantIsEmpty: false,
			wantStart:   3,
			wantEnd:     6,
			wantRealIntervals: []RealIntInterval{
				RealEmpty().Expand(3, 5),
			},
			containsExactIntCases: []containsCase{
				{arg: 2, want: false},
				{arg: 3, want: true},
				{arg: 4, want: true},
				{arg: 5, want: true},
				{arg: 6, want: false},
			},
			expandStartCases: []expandStartCase{
				{args: []int{6}, want: FromStartSizeInt(10, 0, 10)},
				{args: []int{-2}, want: FromStartSizeInt(10, 8, 8)},
				{args: []int{-1}, want: FromStartSizeInt(10, 9, 7)},
				{args: []int{0}, want: FromStartSizeInt(10, 0, 6)},
				{args: []int{0, -1, -2, 3, 4, 5, 0}, want: FromStartSizeInt(10, 8, 8)},
			},
		},
		{
			name:        "90123",
			iv:          FromStartSizeInt(10, 9, 5),
			wantString:  "<mod=10; [9, 9], [0, 3]>",
			wantSize:    5,
			wantIsEmpty: false,
			wantStart:   9,
			wantEnd:     4,
			wantRealIntervals: []RealIntInterval{
				RealEmpty().Expand(9),
				RealEmpty().Expand(0, 1, 2, 3),
			},
			containsExactIntCases: []containsCase{
				{arg: -1, want: false},
				{arg: 0, want: true},
				{arg: 1, want: true},
				{arg: 2, want: true},
				{arg: 3, want: true},
				{arg: 4, want: false},
				{arg: 5, want: false},
				{arg: 6, want: false},
				{arg: 7, want: false},
				{arg: 8, want: false},
				{arg: 9, want: true},
				{arg: 10, want: false},
			},
		},
		{
			name:           "complete 012",
			iv:             FromStartSizeInt(3, 0, 50),
			wantString:     "<mod=3; [0, 2]>",
			wantSize:       3,
			wantIsEmpty:    false,
			wantIsComplete: true,
			wantStart:      0,
			wantEnd:        0,
			wantRealIntervals: []RealIntInterval{
				RealEmpty().Expand(0, 2),
			},
			containsExactIntCases: []containsCase{
				{arg: -1, want: false},
				{arg: 0, want: true},
				{arg: 1, want: true},
				{arg: 2, want: true},
				{arg: 3, want: false},
				{arg: 4, want: false},
			},
		},
		{
			name:        "567 expand start(4, 2)",
			iv:          FromStartSizeInt(10, 5, 3).ExpandStart(4, 2),
			wantString:  "<mod=10; [2, 7]>",
			wantSize:    6,
			wantStart:   2,
			wantEnd:     8,
			wantIsEmpty: false,
			wantRealIntervals: []RealIntInterval{
				RealFromStartSize(2, 6),
			},
		},
	} {
		t.Run(fmt.Sprintf("%s - %s", tt.name, tt.iv.String()), func(t *testing.T) {
			if want, got := tt.wantString, tt.iv.String(); got != want {
				t.Errorf("String() got %q, want %q", got, want)
			}
			if want, got := tt.wantIsEmpty, tt.iv.IsEmpty(); got != want {
				t.Errorf("IsEmpty() got %v, want %v", got, want)
			}
			if want, got := tt.wantIsComplete, tt.iv.IsComplete(); got != want {
				t.Errorf("IsComplete() got %v, want %v", got, want)
			}
			if want, got := tt.wantSize, tt.iv.Size(); got != want {
				t.Errorf("Size() got %v, want %v", got, want)
			}
			if want, got := tt.wantStart, tt.iv.Start(); got != want {
				t.Errorf("Start() got %v, want %v", got, want)
			}
			if want, got := tt.wantEnd, tt.iv.End(); got != want {
				t.Errorf("End() got %v, want %v", got, want)
			}
			if diff := cmp.Diff(tt.wantRealIntervals, tt.iv.RealIntervals(), cmpOpts...); diff != "" {
				t.Errorf("unexpected diff in RealIntervals() (-want +got):\n%s", diff)
			}
			for _, ttt := range tt.containsCases {
				if want, got := ttt.want, tt.iv.Contains(ttt.arg); got != want {
					t.Errorf("Contains(%v) got %v, want %v", ttt.arg, got, want)
				}
			}
			for _, ttt := range tt.containsExactIntCases {
				if want, got := ttt.want, tt.iv.ContainsExactInt(ttt.arg); got != want {
					t.Errorf("ContainsExactInt(%v) got %v, want %v", ttt.arg, got, want)
				}
			}
			for _, ttt := range tt.expandStartCases {
				if want, got := ttt.want, tt.iv.ExpandStart(ttt.args...); !got.EqualSets(want) {
					t.Errorf("%v.ExpandStart(%v) got %v, want %v", tt.iv, ttt.args, got, want)
				}
			}
		})
	}
}

func TestIntIntervalEquality(t *testing.T) {
	for _, tt := range []struct {
		name          string
		a, b          IntInterval
		wantEqualSets bool
	}{
		{
			name:          "empty",
			a:             FromStartSizeInt(10, 0, 0),
			b:             FromStartSizeInt(10, 0, 0),
			wantEqualSets: true,
		},
		{
			name:          "empty different modulus values",
			a:             FromStartSizeInt(10, 0, 0),
			b:             FromStartSizeInt(7, 0, 0),
			wantEqualSets: true,
		},
		{
			name:          "complete different modulus values",
			a:             FromStartSizeInt(10, 0, 10),
			b:             FromStartSizeInt(7, 0, 7),
			wantEqualSets: false,
		},
		{
			name:          "complete different start values",
			a:             FromStartSizeInt(7, 4, 7),
			b:             FromStartSizeInt(7, 0, 7),
			wantEqualSets: true,
		},
		{
			name:          "ExpandStart order for empty set",
			a:             FromStartSizeInt(7, 3, 0).ExpandStart(2),
			b:             FromStartSizeInt(7, 2, 1),
			wantEqualSets: true,
		},
		{
			name:          "ExpandEnd basic",
			a:             FromStartSizeInt(7, 0, 1).ExpandEnd(5),
			b:             FromStartSizeInt(7, 0, 6),
			wantEqualSets: true,
		},
		{
			name:          "ExpandEnd does nothing when passed no args",
			a:             FromStartSizeInt(7, 3, 0).ExpandEnd(),
			b:             FromStartSizeInt(7, 3, 0),
			wantEqualSets: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if want, got := tt.wantEqualSets, tt.a.EqualSets(tt.b); got != want {
				t.Errorf("%s.EqualSets(%s) = %v, want %v", tt.a, tt.b, got, want)
			}
		})
	}
}

func TestRealIntInterval(t *testing.T) {
	type containsCase struct {
		arg  int
		want bool
	}
	type containsIntervalCase struct {
		arg  RealIntInterval
		want bool
	}
	type endCase struct {
		want int
	}
	type expandCase struct {
		values []int
		want   RealIntInterval
	}
	type isEmptyCase struct {
		want bool
	}
	type sizeCase struct {
		want int
	}
	type startCase struct {
		want int
	}

	for _, tt := range []struct {
		name                  string
		iv                    RealIntInterval
		wantString            string
		wantIsEmpty           bool
		wantSize              int
		wantStart             int
		wantEnd               int
		containsCases         []containsCase
		containsIntervalCases []containsIntervalCase
		expandCases           []expandCase
	}{
		{
			name:        "empty",
			iv:          RealEmpty(),
			wantString:  "[empty]",
			wantSize:    0,
			wantIsEmpty: true,
			containsCases: []containsCase{
				{arg: 0, want: false},
				{arg: -1, want: false},
			},
		},
		{
			name:        "345",
			iv:          RealFromStartSize(3, 3),
			wantString:  "[3, 5]",
			wantSize:    3,
			wantIsEmpty: false,
			wantStart:   3,
			wantEnd:     6,
			containsCases: []containsCase{
				{arg: 2, want: false},
				{arg: 3, want: true},
				{arg: 4, want: true},
				{arg: 5, want: true},
				{arg: 6, want: false},
			},
			containsIntervalCases: []containsIntervalCase{
				{RealFromStartSize(2, 2), false},
				{RealFromStartSize(3, 2), true},
				{RealFromStartSize(4, 2), true},
				{RealFromStartSize(5, 2), false},
				{RealFromStartSize(3, 3), true},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s - %s", tt.name, tt.iv.String()), func(t *testing.T) {
			if want, got := tt.wantString, tt.iv.String(); got != want {
				t.Errorf("String() got %q, want %q", got, want)
			}
			if want, got := tt.wantIsEmpty, tt.iv.IsEmpty(); got != want {
				t.Errorf("IsEmpty() got %v, want %v", got, want)
			}
			if want, got := tt.wantSize, tt.iv.Size(); got != want {
				t.Errorf("Size() got %v, want %v", got, want)
			}
			if want, got := tt.wantStart, tt.iv.Start(); got != want {
				t.Errorf("Start() got %v, want %v", got, want)
			}
			if want, got := tt.wantEnd, tt.iv.End(); got != want {
				t.Errorf("End() got %v, want %v", got, want)
			}
			for _, ttt := range tt.containsCases {
				if want, got := ttt.want, tt.iv.Contains(ttt.arg); got != want {
					t.Errorf("Contains(%v) got %v, want %v", ttt.arg, got, want)
				}
			}
			for _, ttt := range tt.containsIntervalCases {
				if want, got := ttt.want, tt.iv.ContainsInterval(ttt.arg); got != want {
					t.Errorf("ContainsInterval(%v) got %v, want %v", ttt.arg, got, want)
				}
			}
		})
	}
}

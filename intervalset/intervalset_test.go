// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package intervalset

import (
	"fmt"
	"reflect"
	"testing"
)

type span struct {
	min, max int
}

// case returns a *span from an Interval interface, or it panics.
func cast(i Interval) *span {
	x, ok := i.(*span)
	if !ok {
		panic(fmt.Errorf("interval must be an span: %v", i))
	}
	return x
}

// zero returns the zero value for span.
func zero() *span {
	return &span{}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *span) String() string {
	return fmt.Sprintf("[%d, %d)", s.min, s.max)
}

func (s *span) Equal(t *span) bool {
	return s.min == t.min && s.max == t.max
}

// Intersect returns the intersection of an interval with another
// interval. The function may panic if the other interval is incompatible.
func (s *span) Intersect(tInt Interval) Interval {
	t := cast(tInt)
	result := &span{
		max(s.min, t.min),
		min(s.max, t.max),
	}
	if result.min < result.max {
		return result
	}
	return zero()
}

// Before returns true if the interval is completely before another interval.
func (s *span) Before(tInt Interval) bool {
	t := cast(tInt)
	return s.max <= t.min
}

// IsZero returns true for the zero value of an interval.
func (s *span) IsZero() bool {
	return s.min == 0 && s.max == 0
}

// Bisect returns two intervals, one on either lower side of x and one on the
// upper side of x, corresponding to the subtraction of x from the original
// interval. The returned intervals are always within the range of the
// original interval.
func (s *span) Bisect(tInt Interval) (Interval, Interval) {
	intersection := cast(s.Intersect(tInt))
	if intersection.IsZero() {
		if s.Before(tInt) {
			return s, zero()
		}
		return zero(), s
	}
	maybeZero := func(min, max int) *span {
		if min == max {
			return zero()
		}
		return &span{min, max}
	}
	return maybeZero(s.min, intersection.min), maybeZero(intersection.max, s.max)

}

// Adjoin returns the union of two intervals, if the intervals are exactly
// adjacent, or the zero interval if they are not.
func (s *span) Adjoin(tInt Interval) Interval {
	t := cast(tInt)
	if s.max == t.min {
		return &span{s.min, t.max}
	}
	if t.max == s.min {
		return &span{t.min, s.max}
	}
	return zero()
}

// Encompass returns an interval that covers the exact extents of two
// intervals.
func (s *span) Encompass(tInt Interval) Interval {
	t := cast(tInt)
	return &span{min(s.min, t.min), max(s.max, t.max)}
}

func TestExtent(t *testing.T) {
	x := &span{20, 40}
	y := &span{60, 100}

	ival := NewSet([]Interval{x, y})
	if got, want := cast(ival.Extent()), (&span{20, 100}); !got.Equal(want) {
		t.Errorf("Extent() = %v, want %v", got, want)
	}
}

func allIntervals(s SetInput) []*span {
	result := []*span{}
	s.IntervalsBetween(s.Extent(), func(x Interval) bool {
		result = append(result, cast(x))
		return true
	})
	return result
}

func TestAdd(t *testing.T) {
	x := NewSet([]Interval{&span{20, 40}})
	y := NewSet([]Interval{&span{60, 111}})

	if got, want := cast(x.Extent()), (&span{20, 40}); !got.Equal(want) {
		t.Errorf("Extent() = %v, want %v", got, want)
	}

	if got, want := cast(y.Extent()), (&span{60, 111}); !got.Equal(want) {
		t.Errorf("Extent() = %v, want %v", got, want)
	}

	x.Add(y)

	if got, want := cast(x.Extent()), (&span{20, 111}); !got.Equal(want) {
		t.Errorf("Extent() = %v, want %v", got, want)
	}

	for _, tt := range []struct {
		name string
		a    *Set
		b    SetInput
		want []*span
	}{
		{
			"empty + empty = empty",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{}),
			[]*span{},
		},
		{
			"empty + [30,111) = [30, 111)",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{&span{30, 111}}),
			[]*span{
				{30, 111},
			},
		},
		{
			"[20, 40) + empty = [20, 40)",
			NewSet([]Interval{&span{20, 40}}),
			NewImmutableSet([]Interval{}),
			[]*span{
				{20, 40},
			},
		},
		{
			"[20, 40) + [60,111)",
			NewSet([]Interval{&span{20, 40}}),
			NewSet([]Interval{&span{60, 111}}),
			[]*span{
				{20, 40},
				{60, 111},
			},
		},
		{
			"[20, 40) + [30,111) = [20, 111)",
			NewSet([]Interval{&span{20, 40}}),
			NewImmutableSet([]Interval{&span{30, 111}}),
			[]*span{
				{20, 111},
			},
		},
	} {
		u := NewImmutableSet(tt.a.AllIntervals()).Union(tt.b)
		tt.a.Add(tt.b)
		if got := allIntervals(tt.a); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
		if got := allIntervals(u); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: [ImmutableSet] got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSub(t *testing.T) {
	for _, tt := range []struct {
		name string
		a    *Set
		b    SetInput
		want []*span
	}{
		{
			"empty - empty = empty",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{}),
			[]*span{},
		},
		{
			"empty - [30,111) = empty",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{&span{30, 111}}),
			[]*span{},
		},
		{
			"[20, 40) - empty = [20, 40)",
			NewSet([]Interval{&span{20, 40}}),
			NewImmutableSet([]Interval{}),
			[]*span{
				{20, 40},
			},
		},
		{
			"[20, 40) - [30,111)",
			NewSet([]Interval{&span{20, 40}}),
			NewSet([]Interval{&span{30, 111}}),
			[]*span{
				{20, 30},
			},
		},
		{
			"[0, 2) [4, 6) [8, 10) - [1, 2) [5, 6) [9, 10)   = [0, 1) [4, 5) [8, 9)",
			NewSet([]Interval{&span{0, 2}, &span{4, 6}, &span{8, 10}}),
			NewSet([]Interval{&span{1, 2}, &span{5, 6}, &span{9, 10}}),
			[]*span{{0, 1}, {4, 5}, {8, 9}},
		},
		{
			"[0...3)[10...13)...[90...93) - all odd numbers",
			func() *Set {
				spans := []Interval{}
				for i := 0; i < 100; i += 10 {
					spans = append(spans, &span{i, i + 3})
				}
				return NewSet(spans)
			}(),
			func() *Set {
				spans := []Interval{}
				for i := 1; i < 100; i += 2 {
					spans = append(spans, &span{i, i + 1})
				}
				return NewSet(spans)
			}(),
			func() []*span {
				spans := []*span{}
				for i := 0; i < 100; i += 10 {
					spans = append(spans, &span{i, i + 1}, &span{i + 2, i + 3})
				}
				return spans
			}(),
		},
	} {
		if got := allIntervals(tt.a.ImmutableSet().Sub(tt.b)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: [ImmutableSet] got %v, want %v", tt.name, got, tt.want)
		}
		tt.a.Sub(tt.b)
		if got := allIntervals(tt.a); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIntersect(t *testing.T) {
	for _, tt := range []struct {
		name string
		a    *Set
		b    SetInput
		want []*span
	}{
		{
			"empty intersect empty = empty",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{}),
			[]*span{},
		},
		{
			"empty intersect [30,111) = empty",
			NewSet([]Interval{}),
			NewImmutableSet([]Interval{&span{30, 111}}),
			[]*span{},
		},
		{
			"[20, 40) intersect empty = empty",
			NewSet([]Interval{&span{20, 40}}),
			NewImmutableSet([]Interval{}),
			[]*span{},
		},
		{
			"[20, 40) intersect [30,111)",
			NewSet([]Interval{&span{20, 40}}),
			NewSet([]Interval{&span{30, 111}}),
			[]*span{{30, 40}},
		},
		{
			"[0, 2) [4, 6) [8, 10) intersect [1, 2) [5, 6) [9, 10)   = [1, 3) [5, 6) [9, 10)",
			NewSet([]Interval{&span{0, 2}, &span{4, 6}, &span{8, 10}}),
			NewSet([]Interval{&span{1, 2}, &span{5, 6}, &span{9, 10}}),
			[]*span{{1, 2}, {5, 6}, {9, 10}},
		},
		{
			"[0, 2) [5, 7) intersect [5, 7)   = [1, 2) [5, 6)",
			// [01...56...]
			// [.12345....]
			NewSet([]Interval{&span{0, 2}, &span{5, 7}}),
			NewSet([]Interval{&span{1, 6}}),
			[]*span{{1, 2}, {5, 6}},
		},
		{
			"[0, 2) [5, 7) intersect [5, 7)   = [1, 2) [5, 6)",
			NewSet([]Interval{&span{1, 6}}),
			NewSet([]Interval{&span{0, 2}, &span{5, 7}}),
			[]*span{{1, 2}, {5, 6}},
		},
		{
			"[0...7)[10...17)...[90...97) intersect (all odd numbers + {4, 14, ... 94})",
			func() *Set {
				spans := []Interval{}
				for i := 0; i < 100; i += 10 {
					spans = append(spans, &span{i, i + 7})
				}
				return NewSet(spans)
			}(),
			func() *Set {
				spans := []Interval{}
				for i := 0; i < 100; i += 10 {
					spans = append(spans, &span{i + 1, i + 2}, &span{i + 3, i + 6}, &span{i + 7, i + 8}, &span{i + 9, i + 10})
				}
				return NewSet(spans)
			}(),
			func() []*span {
				spans := []*span{}
				for i := 0; i < 100; i += 10 {
					spans = append(spans, &span{i + 1, i + 2}, &span{i + 3, i + 6})
				}
				return spans
			}(),
		},
	} {
		if got := allIntervals(tt.a.ImmutableSet().Intersect(tt.b)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: [ImmutableSet] got\n  %v, want\n  %v", tt.name, got, tt.want)
		}
		tt.a.Intersect(tt.b)
		if got := allIntervals(tt.a); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got\n  %v, want\n  %v", tt.name, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  *Set
		elem *span
		want bool
	}{
		{
			name: "{} contains empty interval",
			set:  NewSet([]Interval{}),
			elem: &span{},
			want: true,
		},
		{
			name: "{} does not contain [30,111)",
			set:  NewSet([]Interval{}),
			elem: &span{30, 111},
			want: false,
		},
		{
			name: "[20, 40) contains empty interval",
			set:  NewSet([]Interval{&span{20, 40}}),
			elem: &span{},
			want: true,
		},
		{
			name: "{[0, 5), [10, 15)} contains [0, 5)]",
			set:  NewSet([]Interval{&span{0, 5}, &span{10, 15}}),
			elem: &span{0, 5},
			want: true,
		},
		{
			name: "{[0, 5), [10, 15)} does not contain [0, 6)]",
			set:  NewSet([]Interval{&span{0, 5}, &span{10, 15}}),
			elem: &span{0, 6},
			want: false,
		},
	} {
		if got := tt.set.ImmutableSet().Contains(tt.elem); got != tt.want {
			t.Errorf("%s: [ImmutableSet] set.Contains(%s) = %t, want %t", tt.name, tt.elem, got, tt.want)
		}
		if got := tt.set.Contains(tt.elem); got != tt.want {
			t.Errorf("%s: set.Contains(%s) = %t, want %t", tt.name, tt.elem, got, tt.want)
		}
	}
}

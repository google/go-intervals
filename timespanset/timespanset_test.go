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
package timespanset

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func betweenSlice(set *Set, start, end time.Time) []*timespan {
	ivals := []*timespan{}
	set.IntervalsBetween(start, end, func(s, e time.Time) bool {
		ivals = append(ivals, &timespan{s, e})
		return true
	})
	return ivals
}

func tz() *time.Location {
	x, err := time.LoadLocation("PST8PDT")
	if err != nil {
		panic(fmt.Errorf("timezone not available: %v", err))
	}
	return x
}

var (
	// Dates before and after all the other dates.
	past   = time.Date(1980, time.June, 1, 0, 0, 0, 0, tz())
	future = time.Date(2030, time.June, 1, 0, 0, 0, 0, tz())

	week1 = &timespan{
		time.Date(2015, time.June, 1, 0, 0, 0, 0, tz()),
		time.Date(2015, time.June, 8, 0, 0, 0, 0, tz()),
	}
	week2 = &timespan{
		time.Date(2015, time.June, 8, 0, 0, 0, 0, tz()),
		time.Date(2015, time.June, 15, 0, 0, 0, 0, tz()),
	}
	week3 = &timespan{
		time.Date(2015, time.June, 15, 0, 0, 0, 0, tz()),
		time.Date(2015, time.June, 22, 0, 0, 0, 0, tz()),
	}
)

func weeks1And3() *Set {
	set := Empty()
	set.Insert(week1.start, week1.end)
	set.Insert(week3.start, week3.end)
	return set
}

func weeks123() *Set {
	set := Empty()
	set.Insert(week1.start, week1.end)
	set.Insert(week3.start, week3.end)
	set.Insert(week2.start, week2.end)
	return set
}

func TestIntervalsBetween(t *testing.T) {
	{
		// Test iterating over a single value.
		num := 0
		weeks1And3().IntervalsBetween(past, future, func(_, _ time.Time) bool {
			num++
			return false
		})
		if got, want := num, 1; got != want {
			t.Errorf("want a single result to be returned, got %v", num)
		}
	}
	for _, tt := range []struct {
		name   string
		set    *Set
		bounds *timespan
		want   []*timespan
	}{
		{
			name:   "entire range overlaps with weeks 1 and 2",
			set:    weeks1And3(),
			bounds: &timespan{past, future},
			want:   []*timespan{week1, week3},
		},
		{
			name:   "[week2.start, week3.end] overlaps with week3",
			set:    weeks1And3(),
			bounds: &timespan{week2.start, week3.end},
			want:   []*timespan{week3},
		},
		{
			name:   "zero overlap with week 2",
			set:    weeks1And3(),
			bounds: week2,
			want:   []*timespan{},
		},
		{
			name:   "[week1.start, week3.end] overlaps with week3",
			set:    weeks1And3(),
			bounds: &timespan{week2.start, week3.end},
			want:   []*timespan{week3},
		},
		{
			name:   "weeks123 should be one continuous range",
			set:    weeks123(),
			bounds: &timespan{past, future},
			want:   []*timespan{{week1.start, week3.end}},
		},
	} {
		if got := betweenSlice(tt.set, tt.bounds.start, tt.bounds.end); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: time ranges between %s = %s, want %s", tt.name, tt.bounds, got, tt.want)
		}
	}
}

func TestIntersect(t *testing.T) {
	for _, tt := range []struct {
		name   string
		set    *Set
		bounds *timespan
		want   []*timespan
	}{
		{
			name: "subtract empty from empty",
			set: func() *Set {
				w := Empty()
				w.Sub(Empty())
				return w
			}(),
			bounds: &timespan{past, future},
			want:   []*timespan{},
		},
		{
			name: "subtract weeks 1 and 3 from empty",
			set: func() *Set {
				w := Empty()
				w.Sub(weeks1And3())
				return w
			}(),
			bounds: &timespan{past, future},
			want:   []*timespan{},
		},
		{
			name: "subtract empty from weeks 1 and 3",
			set: func() *Set {
				w := weeks1And3()
				w.Sub(Empty())
				return w
			}(),
			bounds: &timespan{past, future},
			want:   []*timespan{week1, week3},
		},
		{
			name: "subtract weeks 1 and 3 from weeks123",
			set: func() *Set {
				w := weeks123()
				w.Sub(weeks1And3())
				return w
			}(),
			bounds: &timespan{past, future},
			want:   []*timespan{week2},
		},
		{
			name: "subtract eternity from weeks 1 and 3",
			set: func() *Set {
				w := weeks1And3()
				eternity := Empty()
				eternity.Insert(past, future)
				w.Sub(eternity)
				return w
			}(),
			bounds: &timespan{past, future},
			want:   []*timespan{},
		},
	} {
		if got := betweenSlice(tt.set, tt.bounds.start, tt.bounds.end); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: time ranges between %s = %s, want %s", tt.name, tt.bounds, got, tt.want)
		}
	}
}

func TestExtent(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  *Set
		want *timespan
	}{
		{
			name: "weeks123 - weeks13 = week2",
			set: func() *Set {
				w := weeks123()
				w.Sub(weeks1And3())
				return w
			}(),
			want: week2,
		},
		{
			name: "weeks123 - empty set = weeks123",
			set: func() *Set {
				w := weeks123()
				w.Sub(Empty())
				return w
			}(),
			want: &timespan{week1.start, week3.end},
		},
		{
			name: "weeks123 + empty set = weeks123",
			set: func() *Set {
				w := weeks123()
				w.Add(Empty())
				return w
			}(),
			want: &timespan{week1.start, week3.end},
		},
		{
			name: "empty set + weeks123 = weeks123",
			set: func() *Set {
				w := Empty()
				w.Add(weeks123())
				return w
			}(),
			want: &timespan{week1.start, week3.end},
		},
		{
			name: "empty set extent is two zero times",
			set:  Empty(),
			want: &timespan{time.Time{}, time.Time{}},
		},
	} {
		gotStart, gotEnd := tt.set.Extent()

		if got := (&timespan{gotStart, gotEnd}); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: Extent() = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  *Set
		elem *timespan
		want bool
	}{
		{
			name: "weeks 1 and 3 contain week1",
			set:  weeks1And3(),
			elem: week1,
			want: true,
		},
	} {
		if got := tt.set.Contains(tt.elem.start, tt.elem.end); got != tt.want {
			t.Errorf("%s: set.Contains(%s) = %t, want %t", tt.name, tt.elem, got, tt.want)
		}
	}
}

func benchmarkNewSet(numToCreate, numMembers int, overlapping bool, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < numToCreate; i++ {
			set := Empty()
			for j := 0; j < numMembers; j++ {
				if !overlapping {
					set.Insert(week1.start, week1.end)
				} else {
					set.Insert(week1.start.AddDate(0, 0, 7), week1.end.AddDate(0, 0, 7))
				}
			}
		}
	}
}

func TestIntersectCalendar(t *testing.T) {
	for _, tt := range []struct {
		name   string
		set    *Set
		bounds *timespan
		want   []*timespan
	}{
		{
			name: "Intersect: middays August 7-11, 2017",
			set: func() *Set {
				weekdays, _ := weekdaysWeekends(2016, 2018)
				weekdays.Intersect(middays(2016, 2018))
				return weekdays
			}(),
			bounds: &timespan{
				time.Date(2017, time.August, 6, 0, 0, 0, 0, tz()),
				time.Date(2017, time.August, 13, 0, 0, 0, 0, tz()),
			},
			want: []*timespan{
				{
					time.Date(2017, time.August, 7, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 7, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 8, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 8, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 9, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 9, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 10, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 10, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 11, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 11, 13, 0, 0, 0, tz()),
				},
			},
		},
		{
			name: "Sub: middays August 7-9, 2017",
			set: func() *Set {
				weekdays, _ := weekdaysWeekends(2016, 2018)
				weekdays.Sub(middays(2016, 2018))
				return weekdays
			}(),
			bounds: &timespan{
				time.Date(2017, time.August, 6, 0, 0, 0, 0, tz()),
				time.Date(2017, time.August, 9, 0, 0, 0, 0, tz()),
			},
			want: []*timespan{
				{
					time.Date(2017, time.August, 7, 0, 0, 0, 0, tz()),
					time.Date(2017, time.August, 7, 11, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 7, 13, 0, 0, 0, tz()),
					time.Date(2017, time.August, 8, 11, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 8, 13, 0, 0, 0, tz()),
					time.Date(2017, time.August, 9, 0, 0, 0, 0, tz()),
				},
			},
		},
		{
			name: "Intersect: middays August 7-11, 2017",
			set: func() *Set {
				x := middays(2017, 2017)
				weekdays, _ := weekdaysWeekends(2016, 2018)
				x.Intersect(weekdays)
				return x
			}(),
			bounds: &timespan{
				time.Date(2017, time.August, 6, 0, 0, 0, 0, tz()),
				time.Date(2017, time.August, 13, 0, 0, 0, 0, tz()),
			},
			want: []*timespan{
				{
					time.Date(2017, time.August, 7, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 7, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 8, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 8, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 9, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 9, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 10, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 10, 13, 0, 0, 0, tz()),
				},
				{
					time.Date(2017, time.August, 11, 11, 0, 0, 0, tz()),
					time.Date(2017, time.August, 11, 13, 0, 0, 0, tz()),
				},
			},
		},
	} {
		if got := betweenSlice(tt.set, tt.bounds.start, tt.bounds.end); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: time ranges between %s = %s, want %s", tt.name, tt.bounds, got, tt.want)
		}
	}
}

func TestCopyUsingStringEquals(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  *Set
	}{
		{
			name: "3 years worth of weekdays 11am-1pm",
			set: func() *Set {
				weekdays, _ := weekdaysWeekends(2016, 2018)
				weekdays.Intersect(middays(2016, 2018))
				return weekdays
			}(),
		},
	} {
		if got, want := tt.set.Copy().String(), tt.set.String(); got != want {
			t.Errorf("%s: copy = %s != %s", tt.name, got, want)
		}
	}
}

func weekdaysWeekends(yearStart, yearEnd int) (weekdays, weekends *Set) {
	weekdays, weekends = Empty(), Empty()
	jan1 := time.Date(yearStart, time.January, 1, 0, 0, 0, 0, tz())
	for t := jan1; t.Year() <= yearEnd; t = t.AddDate(0, 0, 1) {
		nextDayStart := t.AddDate(0, 0, 1)
		switch t.Weekday() {
		case time.Saturday, time.Sunday:
			weekends.Insert(t, nextDayStart)
		default:
			weekdays.Insert(t, nextDayStart)
		}
	}
	return
}

func middays(yearStart, yearEnd int) *Set {
	middays := Empty()
	jan1 := time.Date(yearStart, time.January, 1, 0, 0, 0, 0, tz())
	for t := jan1; t.Year() <= yearEnd; t = t.AddDate(0, 0, 1) {
		middays.Insert(
			time.Date(t.Year(), t.Month(), t.Day(), 11, 0, 0, 0, tz()),
			time.Date(t.Year(), t.Month(), t.Day(), 13, 0, 0, 0, tz()))
	}
	return middays
}

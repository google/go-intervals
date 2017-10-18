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

package timespanset_test

import (
	"fmt"
	"time"

	"github.com/google/go-intervals/timespanset"
)

func Example() {
	var tz = func() *time.Location {
		x, err := time.LoadLocation("PST8PDT")
		if err != nil {
			panic(fmt.Errorf("timezone not available: %v", err))
		}
		return x
	}()

	type span struct {
		start, end time.Time
	}
	week1 := &span{
		time.Date(2015, time.June, 1, 0, 0, 0, 0, tz),
		time.Date(2015, time.June, 8, 0, 0, 0, 0, tz),
	}
	week2 := &span{
		time.Date(2015, time.June, 8, 0, 0, 0, 0, tz),
		time.Date(2015, time.June, 15, 0, 0, 0, 0, tz),
	}
	week3 := &span{
		time.Date(2015, time.June, 15, 0, 0, 0, 0, tz),
		time.Date(2015, time.June, 22, 0, 0, 0, 0, tz),
	}

	set := timespanset.Empty()
	fmt.Printf("Empty set: %s\n", set)

	set.Insert(week1.start, week3.end)
	fmt.Printf("Week 1-3: %s\n", set)

	set2 := timespanset.Empty()
	set2.Insert(week2.start, week2.end)
	set.Sub(set2)
	fmt.Printf("Week 1-3 minus week 2: %s\n", set)
}

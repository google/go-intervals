# go-intervals

go-intervals is a library for performing set operations on 1-dimensional
intervals, such as time ranges.

Example usage:

```go
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
```

produces

    Empty set: {}
    Week 1-3: {[2015-06-01 00:00:00 -0700 PDT, 2015-06-22 00:00:00 -0700 PDT)}
    Week 1-3 minus week 2: {[2015-06-01 00:00:00 -0700 PDT, 2015-06-08 00:00:00 -0700 PDT), [2015-06-15 00:00:00 -0700 PDT, 2015-06-22 00:00:00 -0700 PDT)}

## Notes

- The intervalset.Set implementation's efficiency could be improved. Insertion
  is best- and worse-case O(n). It could be O(log(n)).

- The library's types and interfaces are still evolving, so expect breaking
  changes.

## Disclaimer

This is not an official Google product.

// Package modinterval provides data structures and functions for working with
// 1-dimensional integer intervals that use modular arithmetic.
//
// See https://fgiesen.wordpress.com/2015/09/24/intervals-in-modular-arithmetic/
// for a discussion of intervals in modular arithmetic.
//
// The core type exported by the library is IntInterval, which is aware of its
// modulus and has several methods for set-like operations, like Contains and
// Expand*.
//
// IntIntervals can also be transformed into standard real-number-based
// intervals; see the RealIntervals() method. An implementation of such
// intervals is provided in this package as well.
package modinterval

import (
	"fmt"
	"strings"
)

// Modulus is a type for working with an integer modulus.
type Modulus int

// Int returns the modulus as a normal integer.
func (m Modulus) Int() int { return int(m) }

// GoModulo returns the result of performing the typical Go modulo operation
// (%) on the argument with m as the modulus. For negative arguments, this
// function may return negative values.
func (m Modulus) GoModulo(a int) int { return a % m.Int() }

// ArrayOffset returns the offset into an array of length m for the given
// position designator according to these rules:
//
// 1) If 0 <= position and position < m, position will be returned as is.
//
// 2) Otherwise, if position > m, (position % m) will be returned.
//
// 3) If position < 0, (position % m) + m will be returned.
//
// Examples:
//
// Modulus(10).ArrayOffset(-1) returns 9.
//
// Modulus(10).ArrayOffset(-11) returns 9.
//
// Modulus(10).ArrayOffset(10) returns 0.
//
// Modulus(10).ArrayOffset(11) returns 1.
//
// Unlike the native Go modulus operator and m.GoModulus(a), ArrayOffset never
// returns negative values.
func (m Modulus) ArrayOffset(position int) int {
	maybeNegative := m.GoModulo(position)
	if maybeNegative < 0 {
		return maybeNegative + m.Int()
	}
	return maybeNegative
}

// IntervalSizeForward returns the size of an modular interval that starts at a
// and ends at b using modulus m.
//
// First a and b are normalized using m.ArrayOffset.
//
// If a == b, 0 is returned. If a < b, b - a is returned. Otherwise, b - a + m
// is returned.
func (m Modulus) IntervalSizeForward(a, b int) int {
	return forwardDistance(m, m.ArrayOffset(a), m.ArrayOffset(b))
}

// IntervalSizeMin returns min(IntervalSizeForward(a, b), IntervalSizeForward(b, a)).
func (m Modulus) IntervalSizeMin(a, b int) int {
	a = m.ArrayOffset(a)
	b = m.ArrayOffset(b)
	return intMin(forwardDistance(m, a, b), forwardDistance(m, b, a))
}

func forwardDistance(m Modulus, a, b int) int {
	if b < a {
		b += m.Int()
	}
	return b - a
}

// IntInterval is an integer interval.
type IntInterval struct {
	modulus Modulus
	start   intPos
	size    intSize
}

// FromStartSizeInt returns an IntInterval from a starting location and a size.
//
// If size > m, size is set to m. If size < 0, FromStartSize panics.
//
// The returned interval is guaranteed to return a Start() value equal to
// m.ArrayOffset(start), even if the size of the interval is 0 or >= m.
func FromStartSizeInt(m Modulus, start, size int) IntInterval {
	if m < 0 {
		panic(fmt.Errorf("invalid modulus = %d is less than 0", m))
	}
	if size > m.Int() {
		size = m.Int()
	} else if size < 0 {
		panic(fmt.Errorf("invalid size = %d is less than 0", size))
	}
	return IntInterval{m, intPos(m.ArrayOffset(start)), intSize(size)}
}

func fromNonemptyStartEnd(m Modulus, start, end int) IntInterval {
	start = m.ArrayOffset(start)
	end = m.ArrayOffset(end)
	if start == end {
		return FromStartSizeInt(m, start, m.Int())
	}
	size := end - start
	if size > 0 {
		return FromStartSizeInt(m, start, size)
	}
	return FromStartSizeInt(m, start, m.Int()-start+end)
}

// String returns a string representation of the interval.
func (iv IntInterval) String() string {
	if iv.IsEmpty() {
		return fmt.Sprintf("<mod=%d; empty>", iv.modulus)
	}
	var parts []string
	for _, part := range iv.RealIntervals() {
		parts = append(parts, part.String())
	}
	return fmt.Sprintf("<mod=%d; %s>", iv.modulus, strings.Join(parts, ", "))
}

// Size returns the number of integers in the interval.
func (iv IntInterval) Size() int {
	return iv.size.int()
}

// Modulus returns the modulus used for the modulo arithmetic assumed by this
// interval.
func (iv IntInterval) Modulus() Modulus {
	return iv.modulus
}

// Start returns the first position in the interval. If the interval is empty,
// the value returned by Start may be non-zero.
func (iv IntInterval) Start() int {
	return iv.start.int() // already normalized
}

// End returns the end position in the interval. If the interval is empty, the
// value returned by End may be non-zero.
//
// End may be less than Start if the interval wraps.
//
// End is equal to start for both the empty set and the complete set.
func (iv IntInterval) End() int {
	return iv.modulus.ArrayOffset(iv.start.int() + iv.size.int())
}

// ExpandStart returns an interval that changes the Start position of the
// interval so that it contains all of the arguments.
//
// Each position designator is transformed by iv.Modulus().ArrayOffset() before
// being considered.
//
// If iv is empty, ExpandStart will use the 'start' parameter passed to
// FromStartSize while expanding the set.
func (iv IntInterval) ExpandStart(positionDesignator ...int) IntInterval {
	if iv.IsComplete() || len(positionDesignator) == 0 {
		return iv
	}

	minStart := iv.start.int()
	for _, val := range positionDesignator {
		offset := iv.Modulus().ArrayOffset(val)
		if offset >= iv.End() {
			// Make offset into a negative number for easier comparison... normalized again later.
			offset -= iv.Modulus().Int()
		}

		minStart = intMin(minStart, offset)
	}

	return FromStartSizeInt(iv.Modulus(), minStart, iv.End()-minStart)
}

// ExpandEnd returns an interval that changes the End position of the
// interval so that it contains all of the arguments.
//
// Each position designator is transformed by iv.Modulus().ArrayOffset() before
// being considered.
//
// If iv is empty, ExpandEnd will use the 'start' parameter passed to
// FromStartSize while expanding the set.
func (iv IntInterval) ExpandEnd(positionDesignator ...int) IntInterval {
	if iv.IsComplete() || len(positionDesignator) == 0 {
		return iv
	}

	origEnd := iv.End()
	maxEnd := origEnd
	for _, val := range positionDesignator {
		minEndToContainPosition := iv.Modulus().ArrayOffset(val + 1)
		if minEndToContainPosition < iv.Start() {
			// Make offset go beyond the allowable normalized length... normalized again later.
			minEndToContainPosition += iv.Modulus().Int()
		}

		maxEnd = intMax(maxEnd, minEndToContainPosition)
	}

	return FromStartSizeInt(iv.Modulus(), iv.Start(), maxEnd-iv.Start())
}

// ExpandMinimal returns a new interval that is expanded the minimal possible
// amount so that it will contain all of its arguments.
//
// Each position designator is transformed by iv.Modulus().ArrayOffset() before
// being considered.
//
// If iv is empty, ExpandMinimal will use the 'start' parameter passed to
// FromStartSize while expanding the set.
func (iv IntInterval) ExpandMinimal(positionDesignator ...int) IntInterval {
	a := iv.ExpandStart(positionDesignator...)
	b := iv.ExpandEnd(positionDesignator...)
	if a.Size() > b.Size() {
		return b
	}
	return a
}

// Contains reports true iff the integer set described by the interval contains
// iv.Modulus().ArrayOffset(positionDesignator).
func (iv IntInterval) Contains(positionDesignator int) bool {
	return iv.ContainsExactInt(iv.Modulus().ArrayOffset(positionDesignator))
}

// ContainsExactInt reports true iff the set described by the interval contains the
// argument. The modulo operation will NOT be applied to the argument.
func (iv IntInterval) ContainsExactInt(i int) bool {
	a, b := iv.realIntervals()
	return a.Contains(i) || b.Contains(i)
}

// realIntervals returns two intervals, either of which may be empty.
//
// The first return value always has the same start as iv and has a maximum
// End() value of iv.modulus - 1.
//
// The second return value always has a start value of 0 (or is empty). The
// second return value will be empty if the first return value has a start value
// of 0.
func (iv IntInterval) realIntervals() (sameStart, zeroStart RealIntInterval) {
	if iv.IsEmpty() {
		return RealIntInterval{}, RealIntInterval{}
	}
	sameStartSize := iv.Size()
	if max := iv.modulus.Int() - iv.Start(); sameStartSize > max {
		sameStartSize = max
	}
	sameStart = RealFromStartSize(iv.Start(), sameStartSize)
	zeroStart = RealFromStartSize(0, iv.Size()-sameStartSize)
	return sameStart, zeroStart
}

// IsEmpty returns true if Size() == 0.
func (iv IntInterval) IsEmpty() bool {
	return iv.size == 0
}

// IsComplete returns true if Size() == iv.Modulus().Int().
func (iv IntInterval) IsComplete() bool {
	return iv.Size() == iv.modulus.Int()
}

// EqualSets returns true if the interval contains exactly the same values as
// another interval. The function ignored the modulus of the two intervals.
func (iv IntInterval) EqualSets(other IntInterval) bool {
	sizesEqual := iv.Size() == other.Size()
	if !sizesEqual {
		return false
	}
	if iv.IsEmpty() {
		return true
	}
	a := iv.normalized()
	b := other.normalized()

	return a.Start() == b.Start()
}

func (iv IntInterval) normalized() IntInterval {
	if !iv.IsComplete() {
		return iv
	}
	return FromStartSizeInt(iv.Modulus(), 0, iv.Size())
}

// RealIntervals returns a set of intervals that together contain exactly the
// same set of integers. The returned slice may be of length 0, 1, or 2.
//
// If the returned slice is of length 1 or 2, the start of the first interval in
// the slice is always equal to iv.Start().
func (iv IntInterval) RealIntervals() []RealIntInterval {
	a, b := iv.realIntervals()
	if a.IsEmpty() && b.IsEmpty() {
		return []RealIntInterval{}
	} else if b.IsEmpty() {
		return []RealIntInterval{a}
	}
	return []RealIntInterval{a, b}
}

// intPos is a position within an interval
type intPos int

func (p intPos) int() int { return int(p) }

// intSize is the size of an interval.
type intSize int

func (p intSize) int() int { return int(p) }

// boundaries describes both sides of an interval's inclusivity.
type boundaries byte

const (
	// [min, max]
	inclusiveInclusive boundaries = iota
	// [min, max)
	inclusiveExclusive
	// (min, max)
	exclusiveExclusive

	// [min, max]
	closedClosed = inclusiveInclusive
	// [min, max)
	closedOpen = inclusiveExclusive
	// (min, max)
	openOpen = exclusiveExclusive
)

func (b boundaries) formatSimpleInterval(min, max string) string {
	switch b {
	case inclusiveInclusive:
		return fmt.Sprintf("[%s, %s]", min, max)
	case inclusiveExclusive:
		return fmt.Sprintf("[%s, %s)", min, max)
	case exclusiveExclusive:
		return fmt.Sprintf("(%s, %s)", min, max)
	default:
		return fmt.Sprintf("<undefined boundaries value %s, %s>", min, max)
	}
}

// RealIntInterval is an integer interval that does not use modular arithmetic.
//
// This type is used by functions of IntInterval that convert a modular interval
// to zero, one, or two non-modular intervals.
type RealIntInterval struct {
	start, size int
}

// RealEmpty returns the empty RealIntInterval.
func RealEmpty() RealIntInterval { return RealIntInterval{} }

// RealFromStartSize returns a non-modular interval from the given start and
// size values.
func RealFromStartSize(start, size int) RealIntInterval {
	return RealIntInterval{start, size}
}

// String returns a string representation of the interval. The empty interval
// returns "[empty]".
func (r RealIntInterval) String() string {
	return r.format(inclusiveInclusive)
}

// format uses the given boundary type to format the interval.
func (r RealIntInterval) format(b boundaries) string {
	if r.IsEmpty() {
		return "[empty]"
	}
	low, high := func() (int, int) {
		switch b {
		case inclusiveInclusive:
			return r.Start(), r.End() - 1
		case inclusiveExclusive:
			return r.Start(), r.End()
		case exclusiveExclusive:
			return r.Start() + 1, r.End()
		default:
			return r.Start(), r.End()
		}
	}()
	return b.formatSimpleInterval(
		fmt.Sprintf("%d", low),
		fmt.Sprintf("%d", high))
}

// IsEmpty reports true iff r.Size() == 0.
func (r RealIntInterval) IsEmpty() bool { return r.Size() == 0 }

// Size returns the number of integers in the interval
func (r RealIntInterval) Size() int { return r.size }

// Contains returns true if the argument is within the interval.
func (r RealIntInterval) Contains(i int) bool {
	return r.Start() <= i && i < r.End()
}

// Intersection returns the intersectino of r and another interval.
func (r RealIntInterval) Intersection(other RealIntInterval) RealIntInterval {
	start := intMax(r.Start(), other.Start())
	end := intMax(r.End(), other.End())
	if end <= start {
		return RealEmpty()
	}
	return RealFromStartSize(start, end-start)
}

// Expand returns an interval that contains the given arguments
func (r RealIntInterval) Expand(value ...int) RealIntInterval {
	if len(value) == 0 {
		return r
	}
	min, max := value[0], value[0]
	if !r.IsEmpty() {
		min, max = r.Start(), r.End()+1
	}
	for _, v := range value {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return RealIntInterval{start: min, size: max - min + 1}
}

// ContainsInterval returns true if the argument is within the interval.
func (r RealIntInterval) ContainsInterval(other RealIntInterval) bool {
	return other.IsEmpty() || (r.Start() <= other.Start() && r.End() >= other.End())
}

// Start returns the inclusive starting position of the interval. The value
// returned is undefined for an empty interval.
func (r RealIntInterval) Start() int {
	return r.start
}

// End returns the exclusive ending position of the interval. The value returned
// is undefined for an empty interval.
func (r RealIntInterval) End() int {
	return r.start + r.size
}

// Add returns an interval shifted in a positive direction by offset.
func (r RealIntInterval) Add(offset int) RealIntInterval {
	return RealFromStartSize(r.Start()+offset, r.Size())
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMinAll(values ...int) int {
	if len(values) == 0 {
		panic("cannot take min of no values")
	}
	best := values[0]
	for _, v := range values {
		if v < best {
			best = v
		}
	}

	return best
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMaxAll(values ...int) int {
	if len(values) == 0 {
		panic("cannot take max of no values")
	}
	best := values[0]
	for _, v := range values {
		if v > best {
			best = v
		}
	}

	return best
}

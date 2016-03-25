//package main

package generate

import (
	"fmt"
	"log"
	"time"

	"github.com/jaybennett89/intmath/intgr"
)

type Position2D struct {
	X float32
	Y float32
}

type Coordinate2D struct {
	X int
	Y int
}

func (c *Coordinate2D) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

func (c *Coordinate2D) GetIndex() int {

	x := c.Y
	y := -c.X
	u := x + y
	v := x - y
	var res int

	if u > 0 {
		if v > 0 {
			x <<= 1
			res = x*(x-1) + v
		} else {
			y <<= 1
			res = y*(y-1) + v
		}
	} else {
		if v < 0 {
			x <<= 1
			res = -x*(1-x) - v
		} else {
			y <<= 1
			res = -y*(1-y) - v
		}
	}
	return res
}

func (c *Coordinate2D) GetFirst() Coordinate2D {
	var first Coordinate2D
	return first
}

func IndexToCoordinate(index int) Coordinate2D {
	var coord Coordinate2D
	coord.X, coord.Y = position(index)
	return coord
}

func first(cycle int) int {
	x := 2*cycle - 1
	return x * x
}

func cycle(index int) int {
	return (intgr.Sqrt(index) + 1) / 2
}

func length(cycle int) int {
	return 8 * cycle
}

func sector(index int) int {
	c := cycle(index)
	offset := index - first(c)
	n := length(c)
	return 4 * offset / n
}

func position(index int) (int, int) {
	c := cycle(index)
	s := sector(index)
	offset := index - first(c) - s*length(c)/4

	switch s {
	case 0:
		return -c, -c + offset + 1
	case 1:
		return -c + offset + 1, c
	case 2:
		return c, c - offset - 1
	default:
		return c - offset - 1, -c
	}
}

func test() {

	loop_size := 500000
	step_size := 1
	step_start := 1

	start_t := time.Now()

	var coord Coordinate2D

	var msg string
	var rev int
	for i := step_start; i < loop_size; i = i + step_size {
		coord = IndexToCoordinate(i)
		msg = coord.String()
		rev = coord.GetIndex()
		fmt.Printf("start=%d, calc=%s, reverse=%d\n", i, msg, rev)
	}

	loop_t := time.Now()

	coord = IndexToCoordinate(2147483647)
	fmt.Printf("max int 32 calc=%s rev=%d\n", coord.String(), coord.GetIndex())

	coord = IndexToCoordinate(9223372036854775000)
	fmt.Printf("max int 64 calc=%s rev=%d\n", coord.String(), coord.GetIndex())

	max_t := time.Now()

	log.Println("testing 10K loop: ", loop_t.Sub(start_t), " int64 max:", max_t.Sub(loop_t))

}

package main

import (
	"math/rand"
	"reflect"
	"testing"
)

var table = []struct {
	A         int
	D         int
	ADice     []int
	DDice     []int
	ExpectedA int
	ExpectedD int
}{
	{4, 2, []int{4, 2, 2}, []int{5, 1}, 3, 1},
	{9, 4, []int{3, 3, 3}, []int{5, 2}, 8, 3},
	{4, 2, []int{1, 1, 1}, []int{5, 1}, 2, 2},
	{6, 1, []int{5, 1, 1}, []int{5}, 5, 1},
	{6, 1, []int{5, 1, 1}, []int{4}, 6, 0},
	{2, 2, []int{4}, []int{5, 1}, 1, 2},
	{2, 3, []int{6}, []int{6, 1}, 1, 3},
	{2, 3, []int{6}, []int{5, 1}, 2, 2},
	{2, 3, []int{1}, []int{6, 1}, 1, 3},
}

func TestGetNewVal(t *testing.T) {

	for idx, test := range table {
		valA, valD := getNewValues(test.A, test.D, test.ADice, test.DDice)

		if valA != test.ExpectedA {
			t.Errorf("Test %v. Got bad value for attacker. Got %v expcected %v", idx, valA, test.ExpectedA)
		}

		if valD != test.ExpectedD {
			t.Errorf("Test %v. Got bad value for defender. Got %v expcected %v", idx, valD, test.ExpectedD)
		}
	}

}

func TestRollDice(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := rand.Intn(3) + 1
		dice := roll_dice(n)
		if len(dice) != n {
			t.Fatalf("len(dice) = %v expected %v", len(dice), n)
		}

		prev := 7
		for _, d := range dice {
			if d > prev {
				t.Fatalf("Dice array %v out of order", dice)
			}
			prev = d
		}
	}
}

var pathTable = []struct {
	Input  string
	Output []int
}{
	{"1", []int{1}},
	{"1,2,3", []int{1, 2, 3}},
	{"1,2,2313,", []int{1, 2, 2313}},
	{"1,2,3,", []int{1, 2, 3}},
	{"1,2,3,,,,,5", []int{1, 2, 3, 5}},
	{",,,1,2,3", []int{1, 2, 3}},
	{"5912391,3", []int{5912391, 3}},
}

func TestParsePath(t *testing.T) {
	for idx, test := range pathTable {
		output := ParsePath(test.Input)
		if !reflect.DeepEqual(output, test.Output) {
			t.Errorf("Test %v not successful. Output %v != %v", idx, output, test.Output)
		}
	}
}

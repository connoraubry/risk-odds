package main

import (
	"log"
	"math/rand"
	"reflect"
	"slices"
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

func BenchmarkSortReverse(b *testing.B) {

	x := 3

	for i := 0; i < b.N; i++ {

		var intSlice []int
		for j := 0; j < x; j++ {
			intSlice = append(intSlice, rand.Intn(6)+1)
		}
		slices.Sort(intSlice)
		slices.Reverse(intSlice)
	}
}

func BenchmarkSortFunction(b *testing.B) {

	x := 3

	for i := 0; i < b.N; i++ {

		var intSlice []int
		for j := 0; j < x; j++ {
			intSlice = append(intSlice, rand.Intn(6)+1)
		}
		slices.SortFunc(intSlice, func(a, b int) int {
			if a > b {
				return -1
			}
			return 1
		})
	}
}

func TestAttackFast(t *testing.T) {
	d_results := make(map[int]int)
	total := 100000
	for i := 0; i < total; i++ {
		new_a, new_d := attack_fast(3, 2)
		if new_a+new_d != 3 {
			log.Fatalf("attack_fast(3, 2) gave extra units back")
		}
		d_results[new_d] += 1
	}

	var resultRange = []struct {
		Low  float64
		High float64
	}{
		{Low: 37, High: 37.4},
		{Low: 33, High: 34},
		{Low: 28.8, High: 29.5},
	}

	for idx, expected := range resultRange {
		actual := 100.0 * (float64(d_results[idx]) / float64(total))

		if actual < expected.Low || actual > expected.High {
			t.Fatalf("Result %v not in range [%v - %v]", actual, expected.Low, expected.High)
		}
	}
}

func BenchmarkAttack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a_val := rand.Intn(3) + 1
		d_val := rand.Intn(2) + 1
		attack(a_val, d_val, roll_dice)
	}
}

func BenchmarkFastAttack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a_val := rand.Intn(3) + 1
		d_val := rand.Intn(2) + 1
		attack_fast(a_val, d_val)
	}
}

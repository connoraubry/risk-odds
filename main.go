package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
)

type Left struct {
	A int
	D int
}

const num_iter int64 = 1000000

var (
	f_mode     = flag.String("mode", "base", "Mode to run tool in. ['base', 'path']")
	f_list     = flag.String("path", "1", "Comma separated path of enemies to go through (Ex. 1; 1,2,1; 1,4,12,3)")
	f_strength = flag.Int("strength", 3, "Strength of units to attack with.")
	f_num      = flag.Int64("num", num_iter, "Iterations to do")
)

func main() {
	fmt.Println("vim-go")

	flag.Parse()

	switch *f_mode {
	case "base":
		fmt.Println("base selected")
		Base()
	case "path":
		fmt.Println("path selected")

		path := ParsePath(*f_list)
		RunPath(*f_strength, path)
	}
}

type PathResult struct {
	AttackResult []map[int]int
	DefendResult []map[int]int
}

func NewPathResult(num int) *PathResult {
	pr := &PathResult{}

	for i := 0; i < num; i++ {
		pr.AttackResult = append(pr.AttackResult, make(map[int]int))
		pr.DefendResult = append(pr.DefendResult, make(map[int]int))
	}
	return pr
}

func PrintMapResult(res map[int]int) {
	var keys []int
	totalVal := 0
	for key, val := range res {
		keys = append(keys, key)
		totalVal += val
	}

	slices.Sort(keys)
	slices.Reverse(keys)

	for _, key := range keys {
		val := res[key]
		percent := (100 * float64(val)) / float64(totalVal)
		fmt.Printf("%v left: (%7d)  %-3.2f%%\n", key, val, percent)
	}
}

func (pr *PathResult) PrettyPrint() {
	fmt.Printf("Results found. using %d iterations.\n", *f_num)
	for i := 0; i < len(pr.AttackResult); i++ {
		fmt.Printf("Path Results for territory %v\n", i)
		fmt.Println("Attacker:")
		PrintMapResult(pr.AttackResult[i])
		fmt.Println("Defender:")
		PrintMapResult(pr.DefendResult[i])
	}
}

func RunPath(strength int, path []int) {

	results := NewPathResult(len(path))

	pb := progressbar.Default(*f_num)

	for i := 0; i < int(*f_num); i++ {
		pb.Add(1)
		attack, defend := SimulateOnePath(strength, path)

		for i := 0; i < len(path); i++ {
			results.AttackResult[i][attack[i]] += 1
			results.DefendResult[i][defend[i]] += 1
		}
	}

	results.PrettyPrint()

}

func SimulateOnePath(strength int, path []int) ([]int, []int) {
	var unitsAtStop = make([]int, len(path))
	var enemyLeft = make([]int, len(path))
	for idx, enemy := range path {
		if strength < 2 {
			unitsAtStop[idx] = strength
			enemyLeft[idx] = enemy

			break
		}

		strength, enemy = attackTilDead(strength, enemy)

		unitsAtStop[idx] = strength
		enemyLeft[idx] = enemy

		strength--

	}
	return unitsAtStop, enemyLeft
}

func ParsePath(path string) []int {
	var pathInt []int
	v := strings.Split(path, ",")
	for idx, entry := range v {
		if entry == "" {
			continue
		}
		v, err := strconv.Atoi(entry)
		if err != nil {
			log.Fatalf("Unable to convert string %v to int. Idx %v", entry, idx)
		}
		pathInt = append(pathInt, v)

	}
	return pathInt
}

func Base() {
	runTest(3, 2)
	runTest(3, 1)
	runTest(2, 2)
	runTest(2, 1)
	runTest(1, 2)
	runTest(1, 1)

}

func runTest(astart, dstart int) {
	test := make(map[Left]int)

	pb := progressbar.Default(*f_num, fmt.Sprintf("Iter Attacker: %v, Defender: %v", astart, dstart))

	for i := 0; i < int(*f_num); i++ {
		a, d := attack(astart, dstart, roll_dice)
		l := Left{A: a, D: d}
		test[l] += 1
		pb.Add(1)
	}

	var keys []Left
	totalValue := 0
	for entry, value := range test {
		keys = append(keys, entry)
		totalValue += value
	}
	slices.SortFunc(keys, func(a, b Left) int {
		if a.A > b.A {
			return 1
		}
		if a.A < b.A {
			return -1
		}
		if a.D > b.D {
			return 1
		}
		if a.D < b.D {
			return -1
		}
		return 0
	})
	for _, key := range keys {
		val := test[key]
		percent := (float64(val) * 100) / float64(totalValue)
		fmt.Printf("%+v: %4v. %-3.2f %%\n", key, val, percent)
	}

}

func getNewValues(attacker, defender int, attackerDice, defenderDice []int) (int, int) {
	for idx := 0; idx < Min(len(defenderDice), len(attackerDice)); idx++ {
		if attackerDice[idx] > defenderDice[idx] {
			defender = defender - 1
		} else {
			attacker = attacker - 1
		}
	}

	return attacker, defender
}

type diceroll_fn func(int) []int

func roll_dice(n int) []int {
	var dice []int
	for i := 0; i < n; i++ {
		dice = append(dice, rand.Intn(6)+1)
	}
	slices.SortFunc(dice, func(a, b int) int {
		if a > b {
			return -1
		}
		return 1
	})
	return dice
}

func attackTilDead(attacking, defending int) (int, int) {

	for attacking > 1 && defending > 0 {
		attack_strength := Min(attacking-1, 3)
		defending_strength := Min(defending, 2)

		a_left, d_left := attack(attack_strength, defending_strength, roll_dice)

		attacking = attacking - attack_strength + a_left
		defending = defending - defending_strength + d_left

	}
	return attacking, defending
}

func attack(attacking, defending int, diceroll diceroll_fn) (int, int) {
	a_dice := diceroll(attacking)
	d_dice := diceroll(defending)

	return getNewValues(attacking, defending, a_dice, d_dice)
}

// returns
func simulate_attack(attacker, defender int) (int, int) {

	for attacker > 1 && defender > 0 {
		numAttackerDice := Min(attacker-1, 3)
		attackerDice := make([]int, numAttackerDice)

		for idx := 0; idx < numAttackerDice; idx++ {
			attackerDice[idx] = rand.Intn(6) + 1
		}

		numDefender := Min(defender, 2)
		defenderDice := make([]int, numDefender)
		for idx := 0; idx < numDefender; idx++ {
			defenderDice[idx] = rand.Intn(6) + 1
		}

		slices.Sort(attackerDice)
		slices.Reverse(attackerDice)
		slices.Sort(defenderDice)
		slices.Reverse(defenderDice)

		attacker, defender = getNewValues(attacker, defender, attackerDice, defenderDice)
	}

	return attacker, defender

}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

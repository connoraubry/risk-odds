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

type diceroll_fn func(int) []int

const num_iter int64 = 1000000

type Odd struct {
	A_loss int
	D_loss int

	Odds float64
}

type Result struct {
	AttackLeft int
	Odds       float64
}

// a -> d -> a_left -> float
var resultTable map[int]map[int]map[int]float64

var oddsLookupTable = map[int]map[int][]Odd{
	1: {
		1: {
			Odd{A_loss: 0, D_loss: 1, Odds: 15.0 / 36.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 21.0 / 36.0},
		},
		2: {
			Odd{A_loss: 0, D_loss: 1, Odds: 55.0 / 216.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 161.0 / 216.0},
		},
	},
	2: {
		1: {
			Odd{A_loss: 0, D_loss: 1, Odds: 125.0 / 216.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 91.0 / 216.0},
		},
		2: {
			Odd{A_loss: 0, D_loss: 2, Odds: 295.0 / 1296.0},
			Odd{A_loss: 1, D_loss: 1, Odds: 420.0 / 1296.0},
			Odd{A_loss: 2, D_loss: 0, Odds: 581.0 / 1296.0},
		},
	},
	3: {
		1: {
			Odd{A_loss: 0, D_loss: 1, Odds: 855.0 / 1296.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 441.0 / 1296.0},
		},
		2: {
			Odd{A_loss: 0, D_loss: 2, Odds: 2890.0 / 7776.0},
			Odd{A_loss: 1, D_loss: 1, Odds: 2611.0 / 7776.0},
			Odd{A_loss: 2, D_loss: 0, Odds: 2275.0 / 7776.0},
		},
	},
}

// func newGetOdds(A_str, D_str int) []Odd {
//
// }

var (
	f_mode       = flag.String("mode", "base", "Mode to run tool in. ['base', 'path', 'sweep', 'test']")
	f_list       = flag.String("path", "1", "Comma separated path of enemies to go through (Ex. 1; 1,2,1; 1,4,12,3)")
	f_strength   = flag.Int("strength", 3, "Strength of units to attack with.")
	f_num        = flag.Int64("num", num_iter, "Iterations to do")
	f_disable_pb = flag.Bool("disable-pb", false, "Disable progress bar")
	f_server     = flag.Bool("server", false, "Run server")
)

func main() {
	flag.Parse()

	if *f_server {
		RunServer()
		return
	}

	switch *f_mode {
	case "base":
		fmt.Println("base selected")
		Base()
	case "path":
		fmt.Println("path selected")
		path := ParsePath(*f_list)
		res := RunPath(*f_strength, path)
		res.PrettyPrint()
	case "sweep":
		fmt.Println("sweep selected")
		path := ParsePath(*f_list)
		Sweep(path)
	case "test":
		t := SingleOdds(2, 1, 1)
		fmt.Println(t)
	case "podds":
		path := ParsePath(*f_list)
		res := PathOdds(*f_strength, path)
		odds := CalculateWinPercent(res)
		fmt.Printf("Success odds: %.2f%%\n", 100*odds)
	}
}

func CalculateWinPercent(v map[int]float64) float64 {
	sum := 0.0
	for val, odd := range v {
		if val > 0 {
			sum += odd
		}
	}
	return sum
}

func init() {
	resultTable = make(map[int]map[int]map[int]float64)
}

func ResLookup(A_start, D_start, A_end int) (float64, bool) {

	if _, ok := resultTable[A_start]; !ok {
		resultTable[A_start] = make(map[int]map[int]float64)
	}
	a_map := resultTable[A_start]
	if _, ok := a_map[D_start]; !ok {
		a_map[D_start] = make(map[int]float64)
	}
	d_map := a_map[D_start]
	result, ok := d_map[A_end]
	return result, ok
}

func ResSave(A_start, D_start, A_end int, value float64) {

	if _, ok := resultTable[A_start]; !ok {
		resultTable[A_start] = make(map[int]map[int]float64)
	}
	a_map := resultTable[A_start]
	if _, ok := a_map[D_start]; !ok {
		a_map[D_start] = make(map[int]float64)
	}
	d_map := a_map[D_start]
	d_map[A_end] = value
}

func PathOdds(A_start int, path []int) map[int]float64 {

	//previous odds are likelihood that we'll get to current node
	prevOdds := make(map[int]float64)
	prevOdds[A_start] = 1

	//run through the path
	for _, entry := range path {
		newOdds := make(map[int]float64)

		//for every starting position in the previous odds,
		// new odds
		for a_start, perct := range prevOdds {
			for a_end, o := range FullOdds(a_start, entry) {
				newOdds[a_end-1] += perct * o
			}
		}
		prevOdds = newOdds
	}

	return prevOdds
}

// with battle A_start vs D_start, what are the odds it
// ends with A_End
func SingleOdds(A_start, D_start, A_End int) float64 {

	if A_End > A_start {
		return 0
	}
	if D_start == 0 {
		if A_End == A_start {
			return 1
		}
		return 0
	}

	//if not ok, it's not saved yet
	if precompute, ok := ResLookup(A_start, D_start, A_End); ok {
		return precompute
	}

	var odds float64 = 0
	for _, result := range DoLookup(A_start, D_start) {
		newA := A_start - result.A_loss
		newD := D_start - result.D_loss
		odds += (result.Odds * SingleOdds(newA, newD, A_End))
	}

	ResSave(A_start, D_start, A_End, odds)
	return odds

}

func FullOdds(A_start, D_start int) map[int]float64 {
	res := make(map[int]float64)

	for i := A_start; i > 0; i-- {
		odds := SingleOdds(A_start, D_start, i)

		//odds greater than 1 in 10 million, add
		if odds > float64(1.0/10000000) {
			res[i] = odds
		}
	}
	return res
}

func OddsTest(A_start, D_start int) {
	var running float64 = 0
	for i := A_start; i > 0; i-- {
		odds := SingleOdds(A_start, D_start, i)

		fmt.Printf("  %v: %v\n", i, odds)
		running += odds
	}
	fmt.Printf("%.4f percent chance of victory\n", 100*running)
}

func DoLookup(a, d int) []Odd {
	new_a := Min(a-1, 3)
	new_d := Min(d, 2)
	return oddsLookupTable[new_a][new_d]
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

func (pr *PathResult) GetSuccessOdds() float64 {
	attackRes := pr.AttackResult[len(pr.AttackResult)-1]
	total := 0
	failOptions := 0
	for key, val := range attackRes {
		total += val
		if key < 2 {
			failOptions += val
		}
	}

	success_rate := total - failOptions

	return (100 * float64(success_rate)) / float64(total)
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

func NewSweep(path []int) map[int]float64 {
	pathSum := 0
	for _, entry := range path {
		pathSum += entry
	}
	sweepStart := len(path) + 1
	sweepEnd := len(path) + (2 * pathSum) + 1
	sweepResults := make(map[int]float64)
	for i := sweepStart; i < sweepEnd; i++ {

		res := PathOdds(i, path)
		odds := CalculateWinPercent(res)
		sweepResults[i] = odds
	}
	return sweepResults
}

func QuietSweep(path []int) {
	pathSum := 0
	for _, entry := range path {
		pathSum += entry
	}
	sweepStart := len(path) + 1
	sweepEnd := len(path) + (2 * pathSum) + 1
	for i := sweepStart; i < sweepEnd; i++ {
		res := RunPathSilent(i, path)
		odds := res.GetSuccessOdds()
		if odds > 99.0 {
			break
		}
	}
}

func Sweep(path []int) {
	pathSum := 0
	for _, entry := range path {
		pathSum += entry
	}
	sweepStart := len(path) + 1
	sweepEnd := len(path) + (2 * pathSum) + 1
	fmt.Println(sweepEnd)

	for i := sweepStart; i < sweepEnd; i++ {
		res := RunPathSilent(i, path)

		odds := res.GetSuccessOdds()
		fmt.Printf("Strength %2d odds: %3.2f%% of success\n", i, odds)
		if odds > 99.0 {
			break
		}
		// if i == sweepStart {
		// 	res.PrettyPrint()
		// }
	}
}
func RunPathSilent(strength int, path []int) *PathResult {

	results := NewPathResult(len(path))

	for i := 0; i < int(*f_num); i++ {
		attack, defend := SimulateOnePath(strength, path, roll_dice)

		for i := 0; i < len(path); i++ {
			results.AttackResult[i][attack[i]] += 1
			results.DefendResult[i][defend[i]] += 1
		}
	}
	return results
}

func RunPath(strength int, path []int) *PathResult {

	results := NewPathResult(len(path))

	pb := progressbar.Default(*f_num)

	for i := 0; i < int(*f_num); i++ {
		pb.Add(1)
		attack, defend := SimulateOnePath(strength, path, roll_dice)

		for i := 0; i < len(path); i++ {
			results.AttackResult[i][attack[i]] += 1
			results.DefendResult[i][defend[i]] += 1
		}
	}
	return results
}

func SimulateOnePath(strength int, path []int, dice_fn diceroll_fn) ([]int, []int) {
	var unitsAtStop = make([]int, len(path))
	var enemyLeft = make([]int, len(path))
	for idx, enemy := range path {
		if strength < 2 {
			unitsAtStop[idx] = strength
			enemyLeft[idx] = enemy

			break
		}

		strength, enemy = attackTilDead(strength, enemy, dice_fn)

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
		a, d := attack_fast(astart, dstart)
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

func roll_dice(n int) []int {
	var dice []int
	for i := 0; i < n; i++ {
		dice = append(dice, rand.Intn(6)+1)
	}
	slices.Sort(dice)
	slices.Reverse(dice)
	return dice
}

func roll_dice_faster(n int) []int {
	dice := make([]int, 3)
	for i := 0; i < n; i++ {
		dice[i] = rand.Intn(6) + 1
	}
	slices.Sort(dice)
	slices.Reverse(dice)
	return dice
}

func attackTilDead(attacking, defending int, dice_fn diceroll_fn) (int, int) {

	for attacking > 1 && defending > 0 {
		attack_strength := Min(attacking-1, 3)
		defending_strength := Min(defending, 2)

		a_left, d_left := attack_fast(attack_strength, defending_strength)

		attacking = attacking - attack_strength + a_left
		defending = defending - defending_strength + d_left

	}
	return attacking, defending
}

func attack_fast(attacking, defending int) (int, int) {
	odds := oddsLookupTable[attacking][defending]

	n := rand.Float64()
	for _, odd := range odds {

		if n < odd.Odds {
			return attacking - odd.A_loss, defending - odd.D_loss
		}
		n -= odd.Odds

	}

	lastOdd := odds[len(odds)-1]

	return attacking - lastOdd.A_loss, defending - lastOdd.D_loss
}
func attack(attacking, defending int, diceroll diceroll_fn) (int, int) {
	a_dice := diceroll(attacking)
	d_dice := diceroll(defending)

	return getNewValues(attacking, defending, a_dice, d_dice)
}

var oddsLookupTableFast = [][][]Odd{
	{},
	{
		{},
		{
			Odd{A_loss: 0, D_loss: 1, Odds: 15.0 / 36.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 21.0 / 36.0},
		},
		{
			Odd{A_loss: 0, D_loss: 1, Odds: 55.0 / 216.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 161.0 / 216.0},
		},
	},
	{
		{},
		{
			Odd{A_loss: 0, D_loss: 1, Odds: 125.0 / 216.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 91.0 / 216.0},
		},
		{
			Odd{A_loss: 0, D_loss: 2, Odds: 295.0 / 1296.0},
			Odd{A_loss: 1, D_loss: 1, Odds: 420.0 / 1296.0},
			Odd{A_loss: 2, D_loss: 0, Odds: 581.0 / 1296.0},
		},
	},
	{
		{},
		{
			Odd{A_loss: 0, D_loss: 1, Odds: 855.0 / 1296.0},
			Odd{A_loss: 1, D_loss: 0, Odds: 441.0 / 1296.0},
		},
		{
			Odd{A_loss: 0, D_loss: 2, Odds: 2890.0 / 7776.0},
			Odd{A_loss: 1, D_loss: 1, Odds: 2611.0 / 7776.0},
			Odd{A_loss: 2, D_loss: 0, Odds: 2275.0 / 7776.0},
		},
	},
}

func attack_fast2(attacking, defending int) (int, int) {
	odds := oddsLookupTableFast[attacking][defending]

	n := rand.Float64()
	for _, odd := range odds {

		if n < odd.Odds {
			return attacking - odd.A_loss, defending - odd.D_loss
		}
		n -= odd.Odds

	}

	lastOdd := odds[len(odds)-1]

	return attacking - lastOdd.A_loss, defending - lastOdd.D_loss
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

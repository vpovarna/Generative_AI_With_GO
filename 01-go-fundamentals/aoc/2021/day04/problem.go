package aoc2021day04

import (
	"flag"
	"fmt"
	"os"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2021, Day4, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2021, Day4, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	return 1
}

func part2(input *string) int {
	return 1
}

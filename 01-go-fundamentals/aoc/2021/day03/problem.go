package aoc2021day03

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative Path for the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2021, Day1, Part1 solution is: %d\n", part1(&input))
}

func part1(input *string) int {
	lines := strings.Split(*input, "\n")

	var gamma, epsilon string

	n := len(lines[0])
	for i := range n {
		countOnes, countZeros := 0, 0

		for _, line := range lines {
			a := string(line[i])
			if a == "0" {
				countZeros += 1
			} else {
				countOnes += 1
			}
		}

		if countOnes > countZeros {
			gamma += "0"
			epsilon += "1"
		} else {
			gamma += "1"
			epsilon += "0"
		}
	}

	e, err := strconv.ParseInt(epsilon, 2, 64)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseInt(gamma, 2, 64)
	if err != nil {
		panic(err)
	}
	return int(e * g)
}

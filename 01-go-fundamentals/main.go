package main

import (
	"fmt"

	validanagram "github.com/povarna/generative-ai-with-go/fundamentals/leetcode/0242_valid_anagram"
)

func main() {
	t := validanagram.IsAnagramWithDict("rat", "cat")
	fmt.Printf("%v\n", t)
}

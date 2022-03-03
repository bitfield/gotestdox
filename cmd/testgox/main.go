package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/bitfield/gotestdox"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		event, err := gotestdox.ParseJSON(scanner.Text())
		if err != nil {
			continue
		}
		if event.Relevant() {
			fmt.Println(event)
		}
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/bitfield/testgox"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		event, err := testgox.ParseJSON(scanner.Text())
		if err != nil {
			continue
		}
		if event.Relevant() {
			fmt.Println(event)
		}
	}
}

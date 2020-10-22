package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func Start(stop chan bool) {
	defer func() {
		stop <- true
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "stop", "quit", "exit":
			fmt.Println("Ok, I quit!")
			return
		default:
			fmt.Println("I dont know what this means: " + text)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

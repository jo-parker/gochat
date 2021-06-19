package main

import (
	"fmt"
)

func one(c1 chan string) {
	for i := 0; i < 10; i++ {
		c1 <- fmt.Sprintf("This is iteration %d in Channel 1", i+1)
	}
	close(c1)
}

func two(c2 chan string) {
	for i := 0; i < 10; i++ {
		c2 <- fmt.Sprintf("This is iteration %d in Channel 2", i+1)
	}
	close(c2)
}

func main() {
	c1 := make(chan string)
	c2 := make(chan string)

	go one(c1)
	go two(c2)

	for {
		select {
			case msg, ok := <- c1:
				fmt.Println(msg)
				if !ok {
					c1 = nil
				}
			case msg, ok := <- c2:
				fmt.Println(msg)
				if !ok {
					c2 = nil
				}
		}

		if c1 == nil && c2 == nil {
			break
		}
	}
}

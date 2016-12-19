package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync"
)

type Result struct {
	URI   string
	Count int
}

func producer() <-chan string {
	// parse URL with homebrew regex - always a bad idea
	re, err := regexp.Compile(`(https?://)?([-\w]+\.)+[-\w]+(/[-\w]+)*/?(\?([-\w]+=[-\w]+(&[-\w]+=[-\w]+)*)?)?(#[-\w]+)?`)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	out := make(chan string)

	go func() {
		defer close(out)

		for scanner.Scan() {
			if s := re.FindString(scanner.Text()); s != "" {
				out <- s
			}
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}()

	return out
}

func consumer(queue <-chan string, concurrency int) <-chan Result {
	pool := make(chan struct{}, concurrency)

	// init pool
	for i := 0; i < concurrency; i++ {
		pool <- struct{}{}
	}

	out := make(chan Result)

	go func() {
		defer close(out)

		var wg sync.WaitGroup
		defer wg.Wait()

		for item := range queue {
			// acquire goroutine
			<-pool

			wg.Add(1)

			go func(item string) {
				defer wg.Done()

				out <- process(item)

				// release goroutine
				pool <- struct{}{}
			}(item)
		}
	}()

	return out
}

func process(s string) Result {
	re, err := regexp.Compile("Go")
	if err != nil {
		panic(err)
	}

	var count int

	if matches := re.FindAllString(s, -1); matches != nil {
		count = len(matches)
	}

	return Result{s, count}
}

const K = 5

func main() {
	pipe := consumer(producer(), K)

	total := 0

	for r := range pipe {
		total += r.Count
		fmt.Printf("Count for %v: %v\n", r.URI, r.Count)
	}

	fmt.Println("Total:", total)
}

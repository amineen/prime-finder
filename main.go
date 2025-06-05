package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func repeatFunc[T any, K any](done <-chan K, fn func() T) <-chan T {
	stream := make(chan T)
	go func() {
		defer close(stream)
		for {
			select {
			case <-done:
				return
			case stream <- fn():
			}
		}
	}()
	return stream
}

func take[T any, K any](done <-chan K, stream <-chan T, n int) <-chan T {

	taken := make(chan T)
	go func() {
		defer close(taken)
		for i := 0; i <= n; i++ {
			select {
			case <-done:
				return
			case taken <- <-stream:
			}
		}
	}()
	return taken
}

// func isPrimeFast(n int) bool {
// 	if n <= 1 {
// 		return false
// 	}
// 	if n <= 3 {
// 		return true
// 	}
// 	if n%2 == 0 || n%3 == 0 {
// 		return false
// 	}
// 	for i := 5; i <= int(math.Sqrt(float64(n))); i += 6 {
// 		if n%i == 0 || n%(i+2) == 0 {
// 			return false
// 		}
// 	}
// 	return true
// }

func primeFinder(done <-chan int, randStream <-chan int) <-chan int {
	isPrime := func(randomInt int) bool {
		for i := randomInt - 1; i > 1; i-- {
			if randomInt%i == 0 {
				return false
			}
		}
		return true
	}

	primes := make(chan int)
	go func() {
		defer close(primes)
		for {
			select {
			case <-done:
				return
			case randomInt := <-randStream:
				if isPrime(randomInt) {
					primes <- randomInt
				}
			}
		}
	}()

	return primes
}

func fanIn[T any](done <-chan int, channels []<-chan T) <-chan T {
	var wg sync.WaitGroup

	fannedInStream := make(chan T)

	transfer := func(c <-chan T) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case fannedInStream <- i:

			}
		}
	}
	for _, c := range channels {
		wg.Add(1)
		go transfer(c)
	}
	go func() {
		wg.Wait()
		close(fannedInStream)
	}()
	return fannedInStream
}

func main() {

	start := time.Now()

	done := make(chan int)
	defer close(done)
	randNumFetcher := func() int { return rand.Intn(500000000) }

	randIntStream := repeatFunc(done, randNumFetcher)

	CPUCount := runtime.NumCPU()

	fmt.Println("CPU counts:", CPUCount)

	//fan out
	primeFinderChannels := make([]<-chan int, CPUCount)

	for i := range CPUCount {
		primeFinderChannels[i] = primeFinder(done, randIntStream)
	}

	//fan in
	fannedInStream := fanIn(done, primeFinderChannels)

	for rando := range take(done, fannedInStream, 10) {
		fmt.Println(rando)
	}

	fmt.Println(time.Since(start))
}

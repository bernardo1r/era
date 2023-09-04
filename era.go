package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const usage = "Usage: %s [OPTIONS...] UPPERBOUND\n" +
	"Calculates the number of prime numbers from [2, UPPERBOUND] using the sieve of Eratosthenes.\n" +
	"UPPERBOUND must be at least 10\n" +
	"\nOptions:\n" +
	"-o    write primes to JSON FILE\n" +
	"-t    number of threads used (default 2)\n"

const chunk = 1 << 17

var nThreads int

var wheel = []int{6, 4, 2, 4, 2, 4, 6, 2}

type Iterator struct {
	alive bool
	max   int
	curr  int
	idx   int
}

func (iter *Iterator) setStart(start int) {
	if start < 7 {
		iter.curr = 1
		return
	}

	mod := start % 30
	iter.curr = start - mod
	if mod <= 2 {
		iter.curr += 1
		iter.idx = 0

	} else if mod <= 10 {
		iter.curr += 7
		iter.idx = 1

	} else if mod <= 12 {
		iter.curr += 11
		iter.idx = 2

	} else if mod <= 16 {
		iter.curr += 13
		iter.idx = 3

	} else if mod <= 18 {
		iter.curr += 17
		iter.idx = 4

	} else if mod <= 22 {
		iter.curr += 19
		iter.idx = 5

	} else if mod <= 28 {
		iter.curr += 23
		iter.idx = 6

	} else {
		iter.curr += 29
		iter.idx = 7
	}
}

func NewWheel(start, max int) *Iterator {
	iter := new(Iterator)
	iter.alive = true
	if max <= 0 {
		iter.max = 0
	} else {
		iter.max = max
	}
	iter.setStart(start)
	return iter
}

func (iter *Iterator) Next() bool {
	if !iter.alive {
		return false
	}

	iter.curr += wheel[iter.idx]
	iter.idx = (iter.idx + 1) % len(wheel)
	if iter.max > 0 && iter.curr >= iter.max {
		iter.alive = false
	}

	return iter.alive
}

func (iter *Iterator) Curr() int {
	if !iter.alive {
		return 0
	}

	return iter.curr
}

func primeEstimative(upperbound int) int {
	if upperbound < 2 {
		return 0
	}

	n := float64(upperbound)
	return int((n / math.Log(n)) * 1.072)
}

func sieveThread(sieve []bool, primes []int, start int, end int) {
	for i := start; i < end; i += chunk {
		jMax := i + chunk
		if jMax > end {
			jMax = end
		}
		for _, p := range primes {
			jIni := max(i-(i%p), p*p)
			if jIni%2 == 0 {
				jIni -= p
			}

			for j := jIni; j < jMax; j += p * 2 {
				sieve[j] = true
			}
		}
	}
}

func sieve(upperbound int) []bool {
	sieve := make([]bool, upperbound+1)
	upperboundSqrt := int(math.Sqrt(float64(upperbound)))
	primes := make([]int, 0, primeEstimative(upperboundSqrt))
	iter := NewWheel(1, upperboundSqrt+1)
	for iter.Next() {
		i := iter.Curr()
		if sieve[i] {
			continue
		}
		primes = append(primes, i)
		for j := i * i; j <= upperboundSqrt; j += i {
			sieve[j] = true
		}
	}

	var wg sync.WaitGroup
	threadChunk := (upperbound - upperboundSqrt) / nThreads
	upperboundSqrt++
	for i := 0; i < nThreads; i++ {
		start := (threadChunk * i) + upperboundSqrt
		end := start + threadChunk
		if i == nThreads-1 {
			end += (upperbound - upperboundSqrt) % nThreads
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			sieveThread(sieve, primes, start, end)
		}()
	}
	wg.Wait()
	return sieve
}

func count(sieve []bool) int {
	var wg sync.WaitGroup
	sums := make([]int, nThreads)
	upperbound := len(sieve) - 1
	threadChunk := upperbound / nThreads
	for i := 0; i < nThreads; i++ {
		i := i
		start := threadChunk * i
		end := start + threadChunk
		if i == nThreads-1 {
			end += (upperbound % nThreads) + 1
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			iter := NewWheel(start, end)
			for iter.Next() {
				if sieve[iter.Curr()] {
					continue
				}
				sums[i]++
			}
		}()
	}
	wg.Wait()
	sum := 3
	for _, v := range sums {
		sum += v
	}
	return sum
}

func writeFile(sieve []bool, filename string) {
	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	var writer *bufio.Writer
	writer = bufio.NewWriter(file)
	writer.WriteString("[2, 3, 5")
	iter := NewWheel(0, len(sieve))
	for iter.Next() {
		i := iter.Curr()
		if sieve[i] {
			continue
		}
		writer.WriteString(", " + strconv.Itoa(i))
	}

	writer.WriteString("]")
	writer.Flush()
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, filepath.Base(os.Args[0]))
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	var outFlag string
	flag.StringVar(&outFlag, "o", "", "write primes to JSON FILE")
	flag.IntVar(&nThreads, "t", 2, "number of threads used")
	flag.Parse()
	if flag.NArg() > 1 {
		log.Fatalf("Too many arguments!\nUpper bound must be specified after flags")
	}

	upperboundStr := flag.Arg(0)
	if upperboundStr == "" {
		log.Fatalln("Upper bound must be specified")
	}

	upperbound, err := strconv.Atoi(upperboundStr)
	if err != nil {
		log.Fatalln("Upper bound must be a number")
	}
	if upperbound < 10 {
		log.Fatalln("Upper bound must be at least 10")
	}

	start := time.Now()
	sieve := sieve(upperbound)
	elapsed := time.Since(start)
	fmt.Printf("Time to make sieve: %v\n", elapsed)
	count := count(sieve)
	fmt.Printf("Primes between 1 and %d: %d\n", upperbound, count)
	if outFlag != "" {
		writeFile(sieve, outFlag)
	}
}

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

const (
	chunk    = 1 << 17
)

var nThreads int

var wheel = []int{6, 4, 2, 4, 2, 4, 6, 2}

type Iterator struct {
	alive bool
	max   int
	curr  int
	idx   int
}

func NewWheel(max int) *Iterator {
	it := new(Iterator)
	if max < 0 {
		return it
	}

	it.alive = true
	it.max = max
	it.curr = 1
	return it
}

func (it *Iterator) Next() bool {
	if !it.alive {
		return false
	}

	it.curr += wheel[it.idx]
	it.idx = (it.idx + 1) % len(wheel)
	if it.curr > it.max {
		it.alive = false
	}

	return it.alive
}

func (it *Iterator) Curr() int {
	if !it.alive {
		return 0
	}

	return it.curr
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
	sieve := make([]bool, int(upperbound+2))
	upperboundSqrt := int(math.Sqrt(float64(upperbound)))
	primes := make([]int, 0, primeEstimative(upperboundSqrt))
	upperbound++
	it := NewWheel(upperboundSqrt)
	for it.Next() {
		i := it.Curr()
		if sieve[i] {
			continue
		}
		primes = append(primes, i)
		for j := i * i; j < upperboundSqrt; j += i {
			sieve[j] = true
		}
	}

	var wg sync.WaitGroup
	threadChunk := (upperbound - upperboundSqrt) / nThreads
	for i := upperboundSqrt; i < upperbound; i += threadChunk {
		start := i
		end := start + threadChunk
		if end > upperbound {
			end = upperbound
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

func writeFile(sieve []bool, upperbound int, file *os.File) int {
	var writer *bufio.Writer
	if file != nil {
		writer = bufio.NewWriter(file)
		writer.WriteString("[2, 3, 5")
	}

	count := 3
	it := NewWheel(upperbound)
	for it.Next() {
		i := it.Curr()
		if sieve[i] {
			continue
		}

		count++
		if file != nil {
			writer.WriteString(", " + strconv.Itoa(i))
		}
	}

	if file != nil {
		writer.WriteString("]")
		writer.Flush()
	}

	return count
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

	var file *os.File
	if outFlag != "" {
		file, err = os.OpenFile(outFlag, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
	}

	start := time.Now()
	sieve := sieve(upperbound)
	elapsed := time.Since(start)
	fmt.Printf("Time to make sieve: %v\n", elapsed)
	count := writeFile(sieve, upperbound, file)
	fmt.Printf("Primes between 1 and %d: %d\n", upperbound, count)
}

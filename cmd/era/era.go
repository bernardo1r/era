package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bernardo1r/era"
)

const usage = "Usage: %s [OPTIONS...] UPPERBOUND\n" +
	"Calculates the number of prime numbers from [2, UPPERBOUND]\n" +
	"using the sieve of Eratosthenes.\n\n" +
	"UPPERBOUND must be at least 10\n" +
	"\nOptions:\n" +
	"  -o  write primes to JSON FILE\n" +
	"  -t  number of threads used (default 2)\n"

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, filepath.Base(os.Args[0]))
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	var (
		outFlag  string
		nThreads int
	)
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
	if nThreads <= 0 {
		log.Fatalln("Number of threads must be greater than zero")
	}
	if nThreads > math.MaxInt {
		log.Fatalln("Number of threads too high")
	}

	start := time.Now()
	sieve := era.Sieve(upperbound, nThreads)
	elapsed := time.Since(start)
	fmt.Printf("Time to make sieve: %v\n", elapsed)
	count := era.Count(sieve, nThreads)
	fmt.Printf("Primes between 1 and %d: %d\n", upperbound, count)
	if outFlag != "" {
		era.WriteFile(sieve, outFlag)
	}
}

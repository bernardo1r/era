// license that can be found in the LICENSE file.
package era_test

import (
	"fmt"

	"github.com/bernardo1r/era"
)

func ExampleNewWheel() {
	iter := era.NewWheel(0, 10)
	iter.Next()
	fmt.Println(iter.Curr())
	// Output: 7
}

func ExampleNewWheel_inRange() {
	iter := era.NewWheel(7, 10)
	iter.Next()
	fmt.Println(iter.Curr())
	// Output: 7
}

func ExampleNewWheel_openRange() {
	iter := era.NewWheel(8, 11)
	for iter.Next() {
		// Won't be printed
		fmt.Println(iter.Curr())
	}
	// Output:
}

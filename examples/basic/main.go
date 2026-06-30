// Example demonstrating f5coefficient package usage.
package main

import (
	"fmt"

	"github.com/0verkilll/f5coefficient"
)

func main() {
	fmt.Println("=== F5 Coefficient Examples ===")
	fmt.Println()

	// Example 1: GetStegoBit - F4 encoding scheme
	fmt.Println("1. GetStegoBit (F4 Encoding):")
	testCoeffs := []int16{3, 4, -3, -4, 1, -1, 2, -2}
	for _, coeff := range testCoeffs {
		bit := f5coefficient.GetStegoBit(coeff)
		sign := "positive"
		if coeff < 0 {
			sign = "negative"
		}
		parity := "even"
		if coeff%2 != 0 {
			parity = "odd"
		}
		fmt.Printf("   Coeff %3d (%s %s) -> bit %d\n", coeff, sign, parity, bit)
	}
	fmt.Println()

	// Example 2: ModifyCoefficient - F5 embedding modification
	fmt.Println("2. ModifyCoefficient (Decrement Absolute Value):")
	modifyTests := []int16{5, -5, 2, -2, 1, -1}
	for _, coeff := range modifyTests {
		modified := f5coefficient.ModifyCoefficient(coeff)
		shrinkage := ""
		if modified == 0 {
			shrinkage = " <- SHRINKAGE!"
		}
		fmt.Printf("   %3d -> %3d%s\n", coeff, modified, shrinkage)
	}
	fmt.Println()

	// Example 3: ReverseModification - Undo F5 modification
	fmt.Println("3. ReverseModification (Undo Embedding):")
	reverseTests := []int16{4, -4, 1, -1, 0}
	for _, coeff := range reverseTests {
		reversed := f5coefficient.ReverseModification(coeff)
		note := ""
		if coeff == 0 {
			note = " (shrinkage case)"
		}
		fmt.Printf("   %3d -> %3d%s\n", coeff, reversed, note)
	}
	fmt.Println()

	// Example 4: IsUsableCoefficient - Filter usable coefficients
	fmt.Println("4. IsUsableCoefficient (Filter for Embedding):")
	usableTests := []struct {
		shuffled int
		coeff    int16
	}{
		{0, 5},   // DC coefficient
		{1, 5},   // Usable AC
		{1, 0},   // Zero coefficient
		{64, -3}, // DC coefficient (block 1)
		{65, -3}, // Usable AC (block 1)
	}
	for _, test := range usableTests {
		usable := f5coefficient.IsUsableCoefficient(test.shuffled, test.coeff)
		reason := ""
		if test.shuffled%64 == 0 {
			reason = " (DC coefficient)"
		} else if test.coeff == 0 {
			reason = " (zero coefficient)"
		}
		fmt.Printf("   shuffled=%3d, coeff=%3d -> usable=%v%s\n",
			test.shuffled, test.coeff, usable, reason)
	}
	fmt.Println()

	// Example 5: IsShrinkageCandidate - Predict shrinkage
	fmt.Println("5. IsShrinkageCandidate (Predict Shrinkage):")
	shrinkageTests := []int16{-2, -1, 0, 1, 2}
	for _, coeff := range shrinkageTests {
		willShrink := f5coefficient.IsShrinkageCandidate(coeff)
		fmt.Printf("   Coeff %3d: will shrink = %v\n", coeff, willShrink)
	}
	fmt.Println()

	// Example 6: ExtractStegoBits - Batch extraction
	fmt.Println("6. ExtractStegoBits (Batch Extraction):")
	coefficients := []int16{3, -2, 5, -7, 4, -4}
	bits := f5coefficient.ExtractStegoBits(coefficients)
	fmt.Printf("   Coefficients: %v\n", coefficients)
	fmt.Printf("   Stego bits:   %v\n", bits)
	fmt.Println()

	// Example 7: Simulated embedding decision
	fmt.Println("7. Embedding Decision Simulation:")
	originalCoeff := int16(5)
	targetBit := 0

	currentBit := f5coefficient.GetStegoBit(originalCoeff)
	fmt.Printf("   Original coefficient: %d (current bit: %d)\n", originalCoeff, currentBit)
	fmt.Printf("   Target bit to embed: %d\n", targetBit)

	if currentBit == targetBit {
		fmt.Println("   Decision: No modification needed")
	} else {
		if f5coefficient.IsShrinkageCandidate(originalCoeff) {
			fmt.Println("   Warning: Modification will cause shrinkage!")
		}
		modified := f5coefficient.ModifyCoefficient(originalCoeff)
		newBit := f5coefficient.GetStegoBit(modified)
		fmt.Printf("   Decision: Modify %d -> %d (new bit: %d)\n",
			originalCoeff, modified, newBit)
	}
}

// Package f5coefficient provides coefficient operations for F5 steganography.
// These operations are shared across f5messageextract, f5messageembed, and f5imagerecover.
package f5coefficient

import (
	"math"

	"github.com/0verkilll/f5core"
)

// Compile-time assertion that BlockSize is 64 (and more generally a power of
// two), which the fast-path `shuffled & (BlockSize-1)` DC check below relies
// on. If the constant is ever changed to a non-power-of-two value, this
// declaration will fail to compile with a non-constant divisor / division-by-
// zero error and the bug will surface during build rather than at runtime.
const _ = 1 / (f5core.BlockSize / 64)
const _ = 1 / (1 - (f5core.BlockSize & (f5core.BlockSize - 1))) // power-of-two guard

// GetStegoBit extracts the steganographic bit value from a DCT coefficient.
//
// F5 uses the F4 sign-based encoding scheme where:
//   - Positive coefficients: LSB directly represents the steganographic value (odd=1, even=0)
//   - Negative coefficients: inverted LSB represents the steganographic value (even=1, odd=0)
//
// This encoding scheme eliminates the statistical weakness of F3 by balancing
// the distribution of steganographic zeros and ones across the coefficient histogram.
//
// Zero coefficients should not be passed to this function as they do not carry
// steganographic data. Use IsUsableCoefficient to filter coefficients first.
//
// Parameters:
//   - coefficient: A non-zero JPEG DCT coefficient in range [-2048, 2047]
//
// Returns:
//   - 0 or 1: The steganographic bit value
//
// Example:
//
//	GetStegoBit(3)  // Returns 1 (positive odd)
//	GetStegoBit(4)  // Returns 0 (positive even)
//	GetStegoBit(-3) // Returns 0 (negative odd)
//	GetStegoBit(-4) // Returns 1 (negative even)
func GetStegoBit(coefficient int16) int {
	bit := int(coefficient) & 1
	if coefficient < 0 {
		return 1 - bit
	}
	return bit
}

// ModifyCoefficient decrements the absolute value of a coefficient.
//
// F5 never overwrites coefficient bits - it only decrements absolute values.
// This reduces statistical detectability compared to LSB overwriting.
//
//   - Positive coefficients: decrement (coeff - 1)
//   - Negative coefficients: increment toward zero (coeff + 1)
//
// This operation may cause shrinkage when |1| or |-1| becomes 0.
// Use IsShrinkageCandidate to check before modifying, or handle
// shrinkage in the embedding loop.
//
// Parameters:
//   - coefficient: A non-zero JPEG DCT coefficient in range [-2048, 2047]
//
// Returns:
//   - The modified coefficient (absolute value decremented by 1)
//
// Example:
//
//	ModifyCoefficient(5)  // Returns 4
//	ModifyCoefficient(-5) // Returns -4
//	ModifyCoefficient(1)  // Returns 0 (shrinkage!)
//	ModifyCoefficient(-1) // Returns 0 (shrinkage!)
func ModifyCoefficient(coefficient int16) int16 {
	if coefficient > 0 {
		return coefficient - 1
	}
	return coefficient + 1
}

// ReverseModification reverses an F5 embedding modification by incrementing
// the absolute value of a coefficient.
//
// During F5 embedding, coefficients are modified by DECREMENTING their absolute values:
//   - Positive coefficients: 5 -> 4 (decrement)
//   - Negative coefficients: -5 -> -4 (increment toward zero)
//
// To recover the original coefficient, we REVERSE this by incrementing absolute values:
//   - Positive coefficients: 4 -> 5 (increment)
//   - Negative coefficients: -4 -> -5 (decrement toward negative infinity)
//
// Zero coefficients are returned unchanged because they represent shrinkage cases
// where the original +/-1 coefficient became 0. Shrinkage recovery is handled
// separately.
//
// Saturation at math.MinInt16:
//
//	The canonical JPEG DCT coefficient range is [-2048, 2047], but callers
//	may pass arbitrary int16 values (the type is wider than the domain).
//	When coefficient == math.MinInt16 (-32768), decrementing would wrap to
//	math.MaxInt16 (+32767) and silently flip the sign — corrupting extraction
//	state. In that one boundary case we return the input unchanged: the
//	operation is undefined for values outside the JPEG domain, and "undefined"
//	is safer than "silent wraparound".
//
// Parameters:
//   - coefficient: A JPEG DCT coefficient value (canonical range: -2048 to 2047)
//
// Returns:
//   - The reversed coefficient with absolute value incremented by 1,
//     the same value if coefficient is 0, or the same value if
//     coefficient == math.MinInt16 (overflow guard).
//
// Example:
//
//	ReverseModification(4)              // Returns 5 (positive increment)
//	ReverseModification(-4)             // Returns -5 (negative decrement)
//	ReverseModification(0)              // Returns 0 (shrinkage case)
//	ReverseModification(1)              // Returns 2
//	ReverseModification(-1)             // Returns -2
//	ReverseModification(math.MinInt16)  // Returns math.MinInt16 (overflow guard)
func ReverseModification(coefficient int16) int16 {
	if coefficient == math.MinInt16 {
		// Cannot safely reverse without wrapping to +32767; return unchanged.
		return coefficient
	}
	if coefficient > 0 {
		return coefficient + 1 // Undo decrement: 4 -> 5
	}
	if coefficient < 0 {
		return coefficient - 1 // Undo increment toward zero: -4 -> -5
	}
	return coefficient // 0 stays 0 (shrinkage case handled separately)
}

// IsUsableCoefficient determines if a coefficient can carry steganographic data.
//
// The F5 algorithm skips two types of coefficients:
//   - DC coefficients: shuffled % BlockSize == 0 (first coefficient of each 8x8 block)
//   - Zero coefficients: cannot represent a bit and would cause ambiguity
//
// Parameters:
//   - shuffled: The permuted index (for DC check)
//   - coefficient: The coefficient value
//
// Returns:
//   - true if the coefficient can be used for embedding/extraction
//   - false if the coefficient should be skipped
//
// Example:
//
//	IsUsableCoefficient(0, 5)   // Returns false (DC coefficient)
//	IsUsableCoefficient(1, 0)   // Returns false (zero coefficient)
//	IsUsableCoefficient(1, 5)   // Returns true (usable AC coefficient)
//	IsUsableCoefficient(64, 5)  // Returns false (DC coefficient)
//	IsUsableCoefficient(65, -3) // Returns true (usable AC coefficient)
func IsUsableCoefficient(shuffled int, coefficient int16) bool {
	// Skip DC coefficients (first coefficient of each 8x8 block).
	// BlockSize is 64 (a power of two, asserted at compile time above), so
	// `shuffled & (BlockSize-1)` is equivalent to `shuffled % BlockSize` on
	// non-negative inputs and ~3x faster in the hot path: no IDIV, pure AND.
	if shuffled&(f5core.BlockSize-1) == 0 {
		return false
	}
	// Skip zero coefficients
	if coefficient == 0 {
		return false
	}
	return true
}

// IsShrinkageCandidate determines if modifying a coefficient will cause shrinkage.
//
// Shrinkage occurs when decrementing |1| or |-1| produces 0. This is a
// fundamental challenge in F5 embedding that requires re-embedding the bit
// using the next available coefficient.
//
// Parameters:
//   - coefficient: The coefficient value
//
// Returns:
//   - true if modifying this coefficient will cause shrinkage
//   - false otherwise
//
// Example:
//
//	IsShrinkageCandidate(1)  // Returns true
//	IsShrinkageCandidate(-1) // Returns true
//	IsShrinkageCandidate(2)  // Returns false
//	IsShrinkageCandidate(-2) // Returns false
//	IsShrinkageCandidate(0)  // Returns false (already zero)
func IsShrinkageCandidate(coefficient int16) bool {
	return coefficient == 1 || coefficient == -1
}

// ExtractStegoBits extracts the steganographic bit values from a slice of coefficients.
//
// This is a helper function that converts raw coefficient values to their
// steganographic bit representation (0 or 1) according to the F4 encoding scheme.
//
// Parameters:
//   - coefficients: The coefficient values
//
// Returns:
//   - A slice of steganographic bits (0 or 1) for each coefficient
//
// Example:
//
//	bits := ExtractStegoBits([]int16{3, -2, 5}) // Returns [1, 1, 1]
//	// 3 is positive odd -> 1
//	// -2 is negative even -> 1
//	// 5 is positive odd -> 1
func ExtractStegoBits(coefficients []int16) []int {
	bits := make([]int, len(coefficients))
	for i, coeff := range coefficients {
		bits[i] = GetStegoBit(coeff)
	}
	return bits
}

// ExtractStegoBitsInto is the zero-allocation variant of ExtractStegoBits. It
// writes the steganographic bit for each coefficient into dst and returns
// dst[:len(coefficients)]. dst is grown (reallocated) only when its capacity
// is too small, so callers that pool/reuse buffers across calls pay no
// allocation cost in the steady state.
//
// This is the hot-path entry point for tools that extract bits for many
// coefficient windows (e.g. F5 bruteforce key recovery where this is called
// once per candidate).
func ExtractStegoBitsInto(dst []int, coefficients []int16) []int {
	if cap(dst) < len(coefficients) {
		dst = make([]int, len(coefficients))
	} else {
		dst = dst[:len(coefficients)]
	}
	if len(coefficients) == 0 {
		return dst
	}
	// Bounds-check elimination hint: we have just re-sliced dst to exactly
	// len(coefficients), so the final in-range index is safe. The compiler
	// uses this to elide per-iteration bounds checks inside the loop.
	_ = dst[len(coefficients)-1]
	for i, coeff := range coefficients {
		dst[i] = GetStegoBit(coeff)
	}
	return dst
}

// NextEmbeddableIndex scans coeffs starting at from and returns the next
// index where IsUsableCoefficient returns true, or -1 if none found.
//
// This is a convenience helper for embedding/extraction loops that need to
// advance over DC positions and zero coefficients. Negative `from` is
// clamped to 0 for defensive convenience; out-of-range `from` (>= len)
// returns -1 immediately.
//
// Parameters:
//   - coeffs: coefficient slice (typically in shuffled order)
//   - from:   starting index (inclusive)
//
// Returns:
//   - The next usable index in [from, len(coeffs)) or -1 if none.
//
// Example:
//
//	idx := NextEmbeddableIndex(coeffs, 0)
//	for idx != -1 {
//	    // process coeffs[idx]
//	    idx = NextEmbeddableIndex(coeffs, idx+1)
//	}
func NextEmbeddableIndex(coeffs []int16, from int) int {
	if from < 0 {
		from = 0
	}
	for i := from; i < len(coeffs); i++ {
		if IsUsableCoefficient(i, coeffs[i]) {
			return i
		}
	}
	return -1
}

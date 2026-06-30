package f5coefficient

import (
	"math"
	"testing"
)

// TestExtractStegoBits_AllZeros covers the pathological "no usable data"
// input. Zero coefficients are not meant to be passed to GetStegoBit per the
// contract, but pipelines that forget to filter must still produce a
// well-defined result (0 for every position) rather than panic.
func TestExtractStegoBits_AllZeros(t *testing.T) {
	coeffs := []int16{0, 0, 0}
	want := []int{0, 0, 0}

	got := ExtractStegoBits(coeffs)
	if len(got) != len(want) {
		t.Fatalf("ExtractStegoBits length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ExtractStegoBits[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// TestReverseModification_Int16MinSaturates exercises the overflow guard
// added to ReverseModification. Input math.MinInt16 (-32768) is outside the
// canonical JPEG coefficient domain but may appear if callers feed raw int16
// data. Naive sign-extending the negative branch (coeff - 1) would wrap to
// +32767 and silently corrupt the stream; the guard clamps to input
// unchanged instead.
func TestReverseModification_Int16MinSaturates(t *testing.T) {
	in := int16(math.MinInt16)
	got := ReverseModification(in)
	if got != in {
		t.Errorf("ReverseModification(math.MinInt16) = %d, want %d (saturated)", got, in)
	}
}

// TestBlockSizeIs64 confirms the assumption baked into the bitmask DC check
// inside IsUsableCoefficient. If f5core.BlockSize is ever changed away from
// 64 the mask-based fast path would silently produce wrong results for all
// inputs between 64 and the new block size, so we assert it explicitly.
func TestBlockSizeIs64(t *testing.T) {
	// Pulled in via coefficient.go's f5core import; re-imported here implicitly
	// through IsUsableCoefficient. We verify against a coefficient we know
	// should be flagged as DC at offset 64.
	if !IsUsableCoefficient(1, 5) {
		t.Fatalf("IsUsableCoefficient(1, 5) = false, want true (AC non-zero)")
	}
	if IsUsableCoefficient(64, 5) {
		t.Fatalf("IsUsableCoefficient(64, 5) = true, want false (DC at block 1)")
	}
	if IsUsableCoefficient(128, 5) {
		t.Fatalf("IsUsableCoefficient(128, 5) = true, want false (DC at block 2)")
	}
	// Confirm the mask ordering: offsets 1..63 inside block 0 must all be
	// usable for non-zero coefficients. If someone swaps in a non-power-of-two
	// BlockSize, these cases would start misclassifying.
	for i := 1; i < 64; i++ {
		if !IsUsableCoefficient(i, 7) {
			t.Errorf("IsUsableCoefficient(%d, 7) = false, want true", i)
		}
	}
}

// TestNextEmbeddableIndex exercises the helper against a hand-built slice
// covering the three skip conditions (DC position, zero coefficient, empty
// tail) as well as the happy path.
func TestNextEmbeddableIndex(t *testing.T) {
	// Lay out 130 coefficients so we cross two 64-wide block boundaries.
	coeffs := make([]int16, 130)
	for i := range coeffs {
		coeffs[i] = 5 // non-zero everywhere by default
	}
	// Force some zeros that must be skipped.
	coeffs[1] = 0
	coeffs[5] = 0
	coeffs[70] = 0

	tests := []struct {
		name string
		from int
		want int
	}{
		// from=0: skip DC at 0, skip zero at 1, land on 2.
		{"skip DC and zero", 0, 2},
		// from=1: zero at 1, land on 2.
		{"skip zero only", 1, 2},
		// from=5: skip zero at 5, land on 6.
		{"skip single zero", 5, 6},
		// from=63: 63 is AC non-zero, return 63.
		{"AC at 63", 63, 63},
		// from=64: DC at 64, land on 65.
		{"skip DC block 1", 64, 65},
		// from=70: zero at 70, land on 71.
		{"skip zero block 1", 70, 71},
		// from=128: DC at 128, land on 129.
		{"skip DC block 2", 128, 129},
		// Negative `from` is clamped to 0.
		{"negative from clamped", -5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextEmbeddableIndex(coeffs, tt.from)
			if got != tt.want {
				t.Errorf("NextEmbeddableIndex(coeffs, %d) = %d, want %d", tt.from, got, tt.want)
			}
		})
	}
}

// TestNextEmbeddableIndex_ExhaustedReturnsMinusOne verifies the sentinel
// contract: scanning past the end returns -1 rather than panicking, so
// embed/extract loops can use `for idx := -1; ...` termination.
func TestNextEmbeddableIndex_ExhaustedReturnsMinusOne(t *testing.T) {
	// All-zero slice has no usable coefficients.
	coeffs := []int16{0, 0, 0, 0}
	if got := NextEmbeddableIndex(coeffs, 0); got != -1 {
		t.Errorf("NextEmbeddableIndex(all zeros, 0) = %d, want -1", got)
	}
	// Out-of-range starting point returns -1 immediately.
	if got := NextEmbeddableIndex(coeffs, 10); got != -1 {
		t.Errorf("NextEmbeddableIndex(_, 10) = %d, want -1 (beyond slice)", got)
	}
	// Empty slice.
	if got := NextEmbeddableIndex(nil, 0); got != -1 {
		t.Errorf("NextEmbeddableIndex(nil, 0) = %d, want -1", got)
	}
}

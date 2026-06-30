package f5coefficient

import (
	"testing"
)

func TestGetStegoBit(t *testing.T) {
	tests := []struct {
		name        string
		coefficient int16
		want        int
	}{
		// Positive odd -> 1
		{"positive odd 1", 1, 1},
		{"positive odd 3", 3, 1},
		{"positive odd 5", 5, 1},
		{"positive odd 7", 7, 1},
		{"positive odd 99", 99, 1},
		{"positive odd 2047", 2047, 1},

		// Positive even -> 0
		{"positive even 2", 2, 0},
		{"positive even 4", 4, 0},
		{"positive even 6", 6, 0},
		{"positive even 100", 100, 0},
		{"positive even 2046", 2046, 0},

		// Negative odd -> 0 (inverted)
		{"negative odd -1", -1, 0},
		{"negative odd -3", -3, 0},
		{"negative odd -5", -5, 0},
		{"negative odd -99", -99, 0},
		{"negative odd -2047", -2047, 0},

		// Negative even -> 1 (inverted)
		{"negative even -2", -2, 1},
		{"negative even -4", -4, 1},
		{"negative even -6", -6, 1},
		{"negative even -100", -100, 1},
		{"negative even -2048", -2048, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStegoBit(tt.coefficient)
			if got != tt.want {
				t.Errorf("GetStegoBit(%d) = %d, want %d", tt.coefficient, got, tt.want)
			}
		})
	}
}

func TestGetStegoBitSymmetry(t *testing.T) {
	// F4 encoding: positive and negative of same absolute value should produce opposite bits
	testCases := []int16{1, 2, 3, 4, 5, 10, 100, 1000}

	for _, absVal := range testCases {
		posBit := GetStegoBit(absVal)
		negBit := GetStegoBit(-absVal)

		// For same absolute value, positive and negative should produce opposite stego bits
		if posBit == negBit {
			t.Errorf("GetStegoBit(%d) = %d, GetStegoBit(%d) = %d, should be opposite",
				absVal, posBit, -absVal, negBit)
		}
	}
}

func TestModifyCoefficient(t *testing.T) {
	tests := []struct {
		name        string
		coefficient int16
		want        int16
	}{
		// Positive coefficients decrement
		{"positive 5", 5, 4},
		{"positive 10", 10, 9},
		{"positive 100", 100, 99},
		{"positive 2047", 2047, 2046},
		{"positive 2 -> 1", 2, 1},
		{"positive 1 -> 0 (shrinkage)", 1, 0},

		// Negative coefficients increment toward zero
		{"negative -5", -5, -4},
		{"negative -10", -10, -9},
		{"negative -100", -100, -99},
		{"negative -2048", -2048, -2047},
		{"negative -2 -> -1", -2, -1},
		{"negative -1 -> 0 (shrinkage)", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModifyCoefficient(tt.coefficient)
			if got != tt.want {
				t.Errorf("ModifyCoefficient(%d) = %d, want %d", tt.coefficient, got, tt.want)
			}
		})
	}
}

func TestModifyCoefficientAbsoluteValueDecreases(t *testing.T) {
	// Verify that ModifyCoefficient always decreases absolute value
	testCases := []int16{2, 3, 10, 100, 1000, -2, -3, -10, -100, -1000}

	for _, coeff := range testCases {
		modified := ModifyCoefficient(coeff)

		absOriginal := coeff
		if absOriginal < 0 {
			absOriginal = -absOriginal
		}

		absModified := modified
		if absModified < 0 {
			absModified = -absModified
		}

		if absModified >= absOriginal {
			t.Errorf("ModifyCoefficient(%d) = %d, absolute value did not decrease", coeff, modified)
		}
	}
}

func TestReverseModification(t *testing.T) {
	tests := []struct {
		name        string
		coefficient int16
		want        int16
	}{
		// Positive coefficients increment
		{"positive 4", 4, 5},
		{"positive 9", 9, 10},
		{"positive 1", 1, 2},
		{"positive 2046", 2046, 2047},

		// Negative coefficients decrement (away from zero)
		{"negative -4", -4, -5},
		{"negative -9", -9, -10},
		{"negative -1", -1, -2},
		{"negative -2047", -2047, -2048},

		// Zero stays zero
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseModification(tt.coefficient)
			if got != tt.want {
				t.Errorf("ReverseModification(%d) = %d, want %d", tt.coefficient, got, tt.want)
			}
		})
	}
}

func TestReverseModificationUndoesModify(t *testing.T) {
	// For non-shrinkage coefficients, Reverse(Modify(x)) should equal x
	testCases := []int16{2, 3, 5, 10, 100, 1000, -2, -3, -5, -10, -100, -1000}

	for _, coeff := range testCases {
		modified := ModifyCoefficient(coeff)
		reversed := ReverseModification(modified)

		if reversed != coeff {
			t.Errorf("ReverseModification(ModifyCoefficient(%d)) = %d, want %d",
				coeff, reversed, coeff)
		}
	}
}

func TestIsUsableCoefficient(t *testing.T) {
	tests := []struct {
		name        string
		shuffled    int
		coefficient int16
		want        bool
	}{
		// DC coefficients (shuffled % 64 == 0) -> not usable
		{"DC block 0", 0, 5, false},
		{"DC block 1", 64, 5, false},
		{"DC block 2", 128, 5, false},
		{"DC block 10", 640, -10, false},

		// Zero coefficients -> not usable
		{"zero coeff", 1, 0, false},
		{"zero coeff other pos", 50, 0, false},

		// AC non-zero coefficients -> usable
		{"AC positive", 1, 5, true},
		{"AC negative", 1, -5, true},
		{"AC position 63", 63, 3, true},
		{"AC position 65", 65, -3, true},
		{"AC large index", 1000, 10, true},

		// Edge cases
		{"position 1 zero", 1, 0, false},
		{"position 127 nonzero", 127, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUsableCoefficient(tt.shuffled, tt.coefficient)
			if got != tt.want {
				t.Errorf("IsUsableCoefficient(%d, %d) = %v, want %v",
					tt.shuffled, tt.coefficient, got, tt.want)
			}
		})
	}
}

func TestIsShrinkageCandidate(t *testing.T) {
	tests := []struct {
		name        string
		coefficient int16
		want        bool
	}{
		// Shrinkage candidates
		{"positive 1", 1, true},
		{"negative -1", -1, true},

		// Non-shrinkage
		{"zero", 0, false},
		{"positive 2", 2, false},
		{"negative -2", -2, false},
		{"positive 100", 100, false},
		{"negative -100", -100, false},
		{"positive max", 2047, false},
		{"negative min", -2048, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsShrinkageCandidate(tt.coefficient)
			if got != tt.want {
				t.Errorf("IsShrinkageCandidate(%d) = %v, want %v",
					tt.coefficient, got, tt.want)
			}
		})
	}
}

func TestExtractStegoBits(t *testing.T) {
	tests := []struct {
		name         string
		coefficients []int16
		want         []int
	}{
		{"empty", []int16{}, []int{}},
		{"single positive odd", []int16{3}, []int{1}},
		{"single negative even", []int16{-2}, []int{1}},
		{"mixed", []int16{3, -2, 5}, []int{1, 1, 1}},
		{"all positive odd", []int16{1, 3, 5, 7}, []int{1, 1, 1, 1}},
		{"all positive even", []int16{2, 4, 6, 8}, []int{0, 0, 0, 0}},
		{"all negative odd", []int16{-1, -3, -5}, []int{0, 0, 0}},
		{"all negative even", []int16{-2, -4, -6}, []int{1, 1, 1}},
		{"complex", []int16{5, -3, 2, -4, 7, -1, 3}, []int{1, 0, 0, 1, 1, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractStegoBits(tt.coefficients)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractStegoBits length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractStegoBits[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetStegoBitReturnsOnlyZeroOrOne(t *testing.T) {
	testCases := []int16{-2048, -1000, -100, -10, -1, 1, 10, 100, 1000, 2047}

	for _, coeff := range testCases {
		bit := GetStegoBit(coeff)
		if bit != 0 && bit != 1 {
			t.Errorf("GetStegoBit(%d) = %d, want 0 or 1", coeff, bit)
		}
	}
}

func BenchmarkGetStegoBit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetStegoBit(int16(i % 2048))
	}
}

func BenchmarkModifyCoefficient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ModifyCoefficient(int16((i % 2046) + 1))
	}
}

func BenchmarkIsUsableCoefficient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsUsableCoefficient(i, int16((i%2046)+1))
	}
}

func BenchmarkExtractStegoBits(b *testing.B) {
	coeffs := make([]int16, 1000)
	for i := range coeffs {
		coeffs[i] = int16((i % 2046) - 1023)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractStegoBits(coeffs)
	}
}

// BenchmarkExtractStegoBitsInto measures the zero-allocation variant with a
// reused output buffer. This matches the workload of F5 bruteforce where the
// same buffer is recycled across millions of candidate evaluations.
func BenchmarkExtractStegoBitsInto(b *testing.B) {
	coeffs := make([]int16, 1000)
	for i := range coeffs {
		coeffs[i] = int16((i % 2046) - 1023)
	}
	buf := make([]int, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = ExtractStegoBitsInto(buf, coeffs)
	}
}

func TestExtractStegoBitsInto_SameResultAsExtractStegoBits(t *testing.T) {
	// Build a set of coefficients covering all four sign/parity cases plus
	// extremes, so the parity assertion exercises the whole table.
	coeffs := []int16{
		0, 1, -1, 2, -2, 3, -3, 4, -4,
		1023, -1023, 1024, -1024,
		32767, -32768,
		99, -99, 100, -100,
	}
	want := ExtractStegoBits(coeffs)

	// First call: nil buffer forces an allocation inside Into.
	got := ExtractStegoBitsInto(nil, coeffs)
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("nil-buf mismatch at %d: coeff=%d got=%d want=%d", i, coeffs[i], got[i], want[i])
		}
	}

	// Second call: pre-allocated buffer is reused.
	buf := make([]int, 0, len(coeffs))
	got = ExtractStegoBitsInto(buf, coeffs)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reused-buf mismatch at %d: coeff=%d got=%d want=%d", i, coeffs[i], got[i], want[i])
		}
	}

	// Third call: undersized buffer triggers reallocation.
	buf = make([]int, 2)
	got = ExtractStegoBitsInto(buf, coeffs)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("grown-buf mismatch at %d: got=%d want=%d", i, got[i], want[i])
		}
	}
}

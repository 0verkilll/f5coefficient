package f5coefficient

import (
	"testing"
)

func FuzzGetStegoBit(f *testing.F) {
	// Add seed corpus
	f.Add(int16(1))
	f.Add(int16(-1))
	f.Add(int16(2))
	f.Add(int16(-2))
	f.Add(int16(100))
	f.Add(int16(-100))
	f.Add(int16(2047))
	f.Add(int16(-2048))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		bit := GetStegoBit(coefficient)

		// Result must be 0 or 1
		if bit != 0 && bit != 1 {
			t.Errorf("GetStegoBit(%d) = %d, want 0 or 1", coefficient, bit)
		}

		// F4 encoding property: positive and negative of same abs value give opposite bits
		if coefficient != 0 && coefficient != -32768 { // Skip edge cases
			posBit := GetStegoBit(coefficient)
			negCoeff := -coefficient
			negBit := GetStegoBit(negCoeff)
			// For F4 encoding, opposite signs should give opposite stego bits
			if posBit == negBit {
				t.Errorf("GetStegoBit(%d)=%d and GetStegoBit(%d)=%d should differ",
					coefficient, posBit, negCoeff, negBit)
			}
		}
	})
}

func FuzzGetStegoBitDeterministic(f *testing.F) {
	f.Add(int16(0))
	f.Add(int16(100))
	f.Add(int16(-100))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		result1 := GetStegoBit(coefficient)
		result2 := GetStegoBit(coefficient)

		if result1 != result2 {
			t.Errorf("GetStegoBit(%d) not deterministic: %d vs %d",
				coefficient, result1, result2)
		}
	})
}

func FuzzModifyCoefficient(f *testing.F) {
	f.Add(int16(1))
	f.Add(int16(-1))
	f.Add(int16(5))
	f.Add(int16(-5))
	f.Add(int16(2047))
	f.Add(int16(-2048))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		if coefficient == 0 {
			return // Skip zero - undefined behavior
		}

		modified := ModifyCoefficient(coefficient)

		// Absolute value should decrease
		absOrig := coefficient
		if absOrig < 0 {
			absOrig = -absOrig
		}
		absMod := modified
		if absMod < 0 {
			absMod = -absMod
		}

		if absMod > absOrig {
			t.Errorf("ModifyCoefficient(%d) = %d, abs value increased",
				coefficient, modified)
		}

		// Sign should be preserved (unless shrinkage to 0)
		if modified != 0 {
			if (coefficient > 0) != (modified > 0) {
				t.Errorf("ModifyCoefficient(%d) = %d, sign changed unexpectedly",
					coefficient, modified)
			}
		}
	})
}

func FuzzReverseModification(f *testing.F) {
	f.Add(int16(0))
	f.Add(int16(1))
	f.Add(int16(-1))
	f.Add(int16(5))
	f.Add(int16(-5))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		reversed := ReverseModification(coefficient)

		// Zero should stay zero
		if coefficient == 0 && reversed != 0 {
			t.Errorf("ReverseModification(0) = %d, want 0", reversed)
		}

		// For non-zero, absolute value should increase
		if coefficient != 0 {
			absOrig := coefficient
			if absOrig < 0 {
				absOrig = -absOrig
			}
			absRev := reversed
			if absRev < 0 {
				absRev = -absRev
			}

			if absRev <= absOrig {
				t.Errorf("ReverseModification(%d) = %d, abs value did not increase",
					coefficient, reversed)
			}
		}
	})
}

func FuzzModifyReverseRoundtrip(f *testing.F) {
	f.Add(int16(2))
	f.Add(int16(-2))
	f.Add(int16(10))
	f.Add(int16(-10))
	f.Add(int16(100))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		// Skip 1 and -1 as they cause shrinkage
		if coefficient == 0 || coefficient == 1 || coefficient == -1 {
			return
		}

		modified := ModifyCoefficient(coefficient)
		reversed := ReverseModification(modified)

		if reversed != coefficient {
			t.Errorf("Roundtrip failed: %d -> %d -> %d",
				coefficient, modified, reversed)
		}
	})
}

func FuzzIsUsableCoefficient(f *testing.F) {
	f.Add(0, int16(5))
	f.Add(1, int16(0))
	f.Add(64, int16(5))
	f.Add(65, int16(5))

	f.Fuzz(func(t *testing.T, shuffled int, coefficient int16) {
		if shuffled < 0 {
			return
		}

		usable := IsUsableCoefficient(shuffled, coefficient)

		// DC coefficients are never usable
		if shuffled%64 == 0 && usable {
			t.Errorf("IsUsableCoefficient(%d, %d) = true, but DC should not be usable",
				shuffled, coefficient)
		}

		// Zero coefficients are never usable
		if coefficient == 0 && usable {
			t.Errorf("IsUsableCoefficient(%d, 0) = true, but zero should not be usable",
				shuffled)
		}

		// AC non-zero should be usable
		if shuffled%64 != 0 && coefficient != 0 && !usable {
			t.Errorf("IsUsableCoefficient(%d, %d) = false, but should be usable",
				shuffled, coefficient)
		}
	})
}

func FuzzIsShrinkageCandidate(f *testing.F) {
	f.Add(int16(0))
	f.Add(int16(1))
	f.Add(int16(-1))
	f.Add(int16(2))
	f.Add(int16(-2))

	f.Fuzz(func(t *testing.T, coefficient int16) {
		isShrinkage := IsShrinkageCandidate(coefficient)

		// Only 1 and -1 are shrinkage candidates
		shouldBeShrinkage := coefficient == 1 || coefficient == -1

		if isShrinkage != shouldBeShrinkage {
			t.Errorf("IsShrinkageCandidate(%d) = %v, want %v",
				coefficient, isShrinkage, shouldBeShrinkage)
		}

		// Verify shrinkage actually happens
		if isShrinkage {
			modified := ModifyCoefficient(coefficient)
			if modified != 0 {
				t.Errorf("Shrinkage candidate %d modified to %d, expected 0",
					coefficient, modified)
			}
		}
	})
}

func FuzzExtractStegoBits(f *testing.F) {
	f.Add([]byte{0, 3, 0, 5}) // Will be interpreted as int16s

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}

		// Convert bytes to int16 slice
		n := len(data) / 2
		coeffs := make([]int16, n)
		for i := 0; i < n; i++ {
			coeffs[i] = int16(data[i*2]) | int16(data[i*2+1])<<8
		}

		bits := ExtractStegoBits(coeffs)

		// Length should match
		if len(bits) != len(coeffs) {
			t.Errorf("ExtractStegoBits returned %d bits for %d coefficients",
				len(bits), len(coeffs))
		}

		// All bits should be 0 or 1
		for i, bit := range bits {
			if bit != 0 && bit != 1 {
				t.Errorf("ExtractStegoBits[%d] = %d, want 0 or 1", i, bit)
			}
		}

		// Each bit should match individual GetStegoBit call
		for i, coeff := range coeffs {
			expected := GetStegoBit(coeff)
			if bits[i] != expected {
				t.Errorf("ExtractStegoBits[%d] = %d, GetStegoBit(%d) = %d",
					i, bits[i], coeff, expected)
			}
		}
	})
}

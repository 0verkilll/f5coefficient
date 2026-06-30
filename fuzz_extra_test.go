package f5coefficient

import (
	"testing"
)

// FuzzGetStegoBitFlipsOnModify asserts the core F5 invariant: modifying a
// coefficient by one step (decrement absolute value) must flip its
// extracted stego bit. If this ever fails, extraction would produce the
// wrong bit after embedding and the algorithm is broken.
//
// The fuzzer skips two cases that are not meaningful under F5 semantics:
//
//   - coefficient == 0: GetStegoBit is undefined for zero; no embedding.
//   - ModifyCoefficient(c) == 0 (shrinkage, |c| == 1): the bit is lost
//     and must be re-embedded elsewhere, so the "flip" invariant does
//     not apply to the shrunk position.
//
// We also skip values outside the canonical JPEG coefficient domain
// [-2048, 2047] because ModifyCoefficient is only contracted on that
// range and wider int16 inputs can change sign via wraparound (distinct
// concern covered by ReverseModification's saturation guard).
func FuzzGetStegoBitFlipsOnModify(f *testing.F) {
	// Seed with a representative sample covering both signs, the shrinkage
	// boundary, mid-range, and the canonical domain extremes.
	seeds := []int16{
		2, 3, 4, 5, 10, 99, 100, 1000, 2047,
		-2, -3, -4, -5, -10, -99, -100, -1000, -2048,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, coefficient int16) {
		// Filter out undefined/out-of-domain inputs (see doc above).
		if coefficient == 0 {
			return
		}
		if coefficient < -2048 || coefficient > 2047 {
			return
		}

		modified := ModifyCoefficient(coefficient)
		if modified == 0 {
			// Shrinkage case: bit is lost, re-embedded elsewhere.
			return
		}

		origBit := GetStegoBit(coefficient)
		modBit := GetStegoBit(modified)

		if origBit == modBit {
			t.Errorf("GetStegoBit did not flip: coefficient=%d bit=%d, "+
				"modified=%d bit=%d (expected flip)",
				coefficient, origBit, modified, modBit)
		}

		// Sanity: both outputs must still be 0 or 1.
		if origBit != 0 && origBit != 1 {
			t.Errorf("GetStegoBit(%d) = %d, want 0 or 1", coefficient, origBit)
		}
		if modBit != 0 && modBit != 1 {
			t.Errorf("GetStegoBit(%d) = %d, want 0 or 1", modified, modBit)
		}
	})
}

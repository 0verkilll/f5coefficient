# Basic Example

This example demonstrates the F4/F5 coefficient operations in the `f5coefficient` package.

## What It Shows

1. **GetStegoBit** - F4 encoding scheme for extracting steganographic bits from coefficients
2. **ModifyCoefficient** - F5 embedding modification (decrement absolute value)
3. **ReverseModification** - Undoing F5 modifications for image recovery
4. **IsUsableCoefficient** - Filtering coefficients suitable for embedding
5. **IsShrinkageCandidate** - Predicting when modification causes shrinkage
6. **ExtractStegoBits** - Batch extraction of steganographic bits
7. **Embedding Decision** - Complete simulation of an embedding decision

## Running

```bash
go run main.go
```

## Expected Output

```
=== F5 Coefficient Examples ===

1. GetStegoBit (F4 Encoding):
   Coeff   3 (positive odd) -> bit 1
   Coeff   4 (positive even) -> bit 0
   Coeff  -3 (negative odd) -> bit 0
   Coeff  -4 (negative even) -> bit 1
   Coeff   1 (positive odd) -> bit 1
   Coeff  -1 (negative odd) -> bit 0
   Coeff   2 (positive even) -> bit 0
   Coeff  -2 (negative even) -> bit 1

2. ModifyCoefficient (Decrement Absolute Value):
     5 ->   4
    -5 ->  -4
     2 ->   1
    -2 ->  -1
     1 ->   0 <- SHRINKAGE!
    -1 ->   0 <- SHRINKAGE!

3. ReverseModification (Undo Embedding):
     4 ->   5
    -4 ->  -5
     1 ->   2
    -1 ->  -2
     0 ->   0 (shrinkage case)

4. IsUsableCoefficient (Filter for Embedding):
   shuffled=  0, coeff=  5 -> usable=false (DC coefficient)
   shuffled=  1, coeff=  5 -> usable=true
   shuffled=  1, coeff=  0 -> usable=false (zero coefficient)
   shuffled= 64, coeff= -3 -> usable=false (DC coefficient)
   shuffled= 65, coeff= -3 -> usable=true

5. IsShrinkageCandidate (Predict Shrinkage):
   Coeff  -2: will shrink = false
   Coeff  -1: will shrink = true
   Coeff   0: will shrink = false
   Coeff   1: will shrink = true
   Coeff   2: will shrink = false

6. ExtractStegoBits (Batch Extraction):
   Coefficients: [3 -2 5 -7 4 -4]
   Stego bits:   [1 1 1 0 0 1]

7. Embedding Decision Simulation:
   Original coefficient: 5 (current bit: 1)
   Target bit to embed: 0
   Decision: Modify 5 -> 4 (new bit: 0)
```

## Key Concepts

### F4 Encoding Scheme

The F4 scheme maps coefficient values to steganographic bits:

| Sign | Parity | Stego Bit |
|------|--------|-----------|
| Positive | Odd | 1 |
| Positive | Even | 0 |
| Negative | Odd | 0 |
| Negative | Even | 1 |

This ensures that positive and negative coefficients of the same absolute
value encode opposite bits, which helps maintain statistical properties.

### Shrinkage

Shrinkage occurs when modifying a coefficient with absolute value 1:
- `1 -> 0` (positive shrinkage)
- `-1 -> 0` (negative shrinkage)

When shrinkage occurs, the coefficient becomes zero and can no longer
carry steganographic data. F5 handles this by re-embedding the data
in the next available coefficient.

### Usable Coefficients

Not all coefficients can be used for embedding:
- **DC coefficients** (position 0 in each 8x8 block) are skipped
- **Zero coefficients** cannot carry data (already at minimum value)

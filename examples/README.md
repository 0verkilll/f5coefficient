# Examples

This directory contains examples demonstrating how to use the `f5coefficient` package.

## Available Examples

| Example | Description |
|---------|-------------|
| [basic](./basic/) | F4 coefficient operations including bit extraction, modification, and shrinkage detection |

## Running Examples

Each example can be run directly with `go run`:

```bash
cd examples/basic
go run main.go
```

Or run from the repository root:

```bash
go run ./examples/basic
```

## Example Output

Running the basic example produces output like:

```
=== F5 Coefficient Examples ===

1. GetStegoBit (F4 Encoding):
   Coeff   3 (positive odd) -> bit 1
   Coeff   4 (positive even) -> bit 0
   Coeff  -3 (negative odd) -> bit 0
   Coeff  -4 (negative even) -> bit 1
   ...

2. ModifyCoefficient (Decrement Absolute Value):
     5 ->   4
    -5 ->  -4
     1 ->   0 <- SHRINKAGE!
   ...
```

## Creating New Examples

When adding new examples:

1. Create a new directory under `examples/`
2. Add a `main.go` file with a `main` function
3. Add a `go.mod` file with appropriate dependencies
4. Add a `README.md` describing the example
5. Update this README to include the new example in the table

# Hashing Strategies and Benchmarks

This directory contains the implementations for various hashing algorithms used in the URL shortener. We use both cryptographic and non-cryptographic hashes to demonstrate the performance differences between them.

## Cryptographic vs Non-Cryptographic Hashes

*   **Cryptographic Hashes (`MD5`, `SHA-256`):** Designed for security. They intentionally use complex mathematics and multiple rounds of processing to prevent brute-forcing and collisions. Because of this, they consume significant CPU cycles.
*   **Non-Cryptographic Hashes (`FNV`, `Murmur3`, `XXHash`):** Designed for speed and uniform distribution. They use fast bitwise math to scramble data quickly. They do not care about security (preventing reverse engineering).

**Why use non-cryptographic hashes for URL shorteners?** 
In a URL shortener, reversing a hash just gives you the original public URL—there are no secrets being protected. Because security is not a concern, spending extra CPU cycles on cryptographic math is a waste of server resources. Using a non-cryptographic hash (like XXHash) allows a smaller server to handle significantly more traffic.

## Running the Benchmarks

You can run the benchmarks locally to see the performance differences on your specific hardware.

Navigate to this directory and run:
```bash
go test -bench=.
```

## How to Read the Benchmark Results

When you run the benchmark, you will get output similar to this:

```text
BenchmarkXXHash-12      30861286                38.32 ns/op          104 B/op          2 allocs/op
```

*   **`BenchmarkXXHash-12`**: The name of the benchmark. The `-12` indicates the number of CPU threads used (e.g., 12 cores).
*   **`30861286` (Total Iterations)**: How many times the function was executed during the test. **Higher is better.**
*   **`38.32 ns/op` (Nanoseconds per operation)**: How long a single hash operation took. **Lower is better.** This is the primary speed metric.
*   **`104 B/op` (Bytes allocated)**: Memory allocated on the heap per operation. **Lower is better.**
*   **`2 allocs/op` (Allocations)**: Number of times memory had to be allocated during a single operation. **Lower is better.**

## Key Takeaways from Benchmarks

1. **XXHash / Murmur are the Speed Champions:** They are significantly faster than MD5, making them the ideal choice for high-throughput systems like a URL shortener.
2. **Hardware Acceleration (Apple Silicon Quirk):** On newer processors like the Apple M-series chips, you might notice `SHA-256` running faster than `MD5`. This is because Apple Silicon (and some other modern CPUs) have dedicated hardware circuits physically built into the chip exclusively to compute SHA-256 instantly. Even with this massive hardware advantage, `XXHash` still outperforms it!
3. **FNV is a Memory Champion:** FNV allocates extremely little memory (often single-digit bytes), making it useful if memory constraints are stricter than CPU constraints.

> [!NOTE]
> *The benchmark metrics shown above were demonstrated on an **Apple M4 Pro with 24 GB RAM**. These results (especially the extremely fast SHA-256 times) will vary significantly on different machine architectures and older Intel-based processors.*

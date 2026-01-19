# BMSSP: Breaking the Sorting Barrier for Directed Single-Source Shortest Paths

[![Go Reference](https://pkg.go.dev/badge/github.com/deicod/bmssp.svg)](https://pkg.go.dev/github.com/deicod/bmssp)
![CI](https://github.com/deicod/bmssp/workflows/Go/badge.svg)

`bmssp` is a Go implementation of the **Bounded Multi-Source Shortest Path (BMSSP)** algorithm, as described in the paper:

> **Breaking the Sorting Barrier for Directed Single-Source Shortest Paths**  
> *Ran Duan, Jiayi Mao, Xiao Mao, Xinkai Shu, Longhui Yin*  
> [arXiv:2504.17033](https://arxiv.org/abs/2504.17033)

This algorithm solves the Single-Source Shortest Path (SSSP) problem on directed graphs with non-negative real edge weights in $O(m \log^{2/3} n)$ time, surpassing the classic Dijkstra limit of $O(m + n \log n)$ for sparse graphs.

## Features

- **Advanced Algorithm**: Implements the recursive BMSSP structure with pivot selection as described in the paper.
- **Sparse Graph Optimization**: Designed to outperform standard Dijkstra on large, sparse graphs.
- **Gonum Integration**: Includes an adapter for the popular [gonum/graph](https://github.com/gonum/gonum) library, allowing easy integration with existing Go implementations.
- **Hybrid Approach**: Automatically falls back to highly optimized Dijkstra for small graphs ($N < 1000$) to minimize overhead.

## Installation

```bash
go get github.com/deicod/bmssp
```

## Usage

### Basic Usage

Use the native `bmssp.Graph` for maximum performance with dense integer indices.

```go
package main

import (
	"fmt"
	"github.com/deicod/bmssp"
)

func main() {
	// Create a graph with 5 vertices
	g := bmssp.NewGraph(5)
	
	// Add edges: u -> v (weight)
	g.AddEdge(0, 1, 4.0)
	g.AddEdge(0, 2, 2.0)
	g.AddEdge(1, 2, 5.0)
	g.AddEdge(1, 3, 10.0)
	g.AddEdge(2, 3, 3.0)
	g.AddEdge(3, 4, 1.0)

	// Create solver
	solver := bmssp.NewSolver(g)
	
	// Find shortest path from vertex 0 to 4
	dist, path := solver.Solve(0, 4)

	fmt.Printf("Shortest Distance: %f\n", dist)
	fmt.Printf("Path: %v\n", path)
}
```

### Using with Gonum

If you are using `gonum/graph`, you can use the built-in adapter.

```go
package main

import (
	"fmt"
	"log"

	"github.com/deicod/bmssp"
	"gonum.org/v1/gonum/graph/simple"
)

func main() {
	// Create a simple weighted directed graph
	g := simple.NewWeightedDirectedGraph(0, 0)
	
	n1 := g.NewNode()
	n2 := g.NewNode()
	g.AddNode(n1)
	g.AddNode(n2)
	
	// Add weighted edge
	g.SetWeightedEdge(g.NewWeightedEdge(n1, n2, 42.0))

	// Solve using BMSSP
	dist, path, err := bmssp.SolveGonum(g, n1.ID(), n2.ID())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Distance: %f\n", dist)
	fmt.Printf("Path IDs: %v\n", path)
}
```

## Complexity

The paper's algorithm achieves a time complexity of $O(m \log^{2/3} n)$ in the comparison-addition model.
This implementation follows the block-based frontier from Lemma 3.3 (with sorted blocks and batch
prepends) and maintains a lightweight balanced-tree index (treap) for block bounds.

- **$n$**: Number of vertices
- **$m$**: Number of edges

This is achieved by avoiding the global sorting bottleneck of Dijkstra's algorithm (using a global priority queue). Instead, BMSSP maintains vertices in approximate order using block-based buckets and sorts them only when necessary ("lazy sorting") within recursive subproblems defined by selected "pivot" vertices.

## Implementation Details

The implementation follows the paper's structure:
1.  **Pivot Selection**: Identifies strategic vertices to partition the search space.
2.  **Frontier Management**: Uses the block-based D0/D1 structure with batch prepends and block pulls.
3.  **Recursive Solving**: A divide-and-conquer approach that solves subproblems with tighter distance bounds.
4.  **Constant-Degree Reduction**: Applies the cycle-based transformation from §2 for BMSSP runs and maps
    paths back to original vertices.

For small graphs or base cases of the recursion, a bounded version of Dijkstra's algorithm is used.

## License

MIT

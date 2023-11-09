package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// DEBUG stuff

const enable_logging = false

type RecursiveLogger struct {
	depth int
}

func (l *RecursiveLogger) Printf(format string, args ...any) {
	if enable_logging {
		msg := fmt.Sprintf(format, args...)
		repl := strings.Repeat("\t", l.depth)
		fmt.Fprint(os.Stderr, repl+msg)
	}
}

func (l *RecursiveLogger) StepDown() {
	l.depth++
}

func (l *RecursiveLogger) StepUp() {
	l.depth--
	if l.depth < 0 {
		panic("invalid step up")
	}
}

func Nodes2String(nodes ...*Node) string {
	var s strings.Builder
	s.WriteString("[ ")
	for _, n := range nodes {
		s.WriteString(fmt.Sprintf("%v ", n.id))
	}
	s.WriteString("]")
	return s.String()
}

var (
	logger = &RecursiveLogger{0}
)

// algorithm stuff
// Priority Queue implementation from https://golang.org/pkg/container/heap/
type MaxPriorityQueue []*Node

func (pq MaxPriorityQueue) Len() int {
	return len(pq)
}

func (pq MaxPriorityQueue) Less(i, j int) bool {
	a := pq[i].covered
	b := pq[j].covered
	c := pq[i].edgeCount > pq[j].edgeCount
	return (!a && b) || (a == b && c)
}

func (pq MaxPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index_max = i
	pq[j].index_max = j
}

func (pq *MaxPriorityQueue) Push(x any) {
	n := len(*pq)
	node := x.(*Node)
	node.index_max = n
	*pq = append(*pq, node)
}

func (pq *MaxPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil      // avoid memory leak
	node.index_max = -1 // for safety
	*pq = old[0 : n-1]
	return node
}

func (pq *MaxPriorityQueue) update(node *Node, edgeCount int, covered bool) {
	node.edgeCount = edgeCount
	node.covered = covered
	heap.Fix(pq, node.index_max)
}

// Priority Queue implementation from https://golang.org/pkg/container/heap/
type MinPriorityQueue []*Node

func (pq MinPriorityQueue) Len() int {
	return len(pq)
}

func (pq MinPriorityQueue) Less(i, j int) bool {
	a := pq[i].covered
	b := pq[j].covered
	c := pq[i].edgeCount < pq[j].edgeCount
	return (!a && b) || (a == b && c)
}

func (pq MinPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index_min = i
	pq[j].index_min = j
}

func (pq *MinPriorityQueue) Push(x any) {
	node := x.(*Node)
	node.index_min = len(*pq)
	*pq = append(*pq, node)
}

func (pq *MinPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil      // avoid memory leak
	node.index_min = -1 // for safety
	*pq = old[0 : n-1]
	return node
}

func (pq *MinPriorityQueue) update(node *Node, edgeCount int, covered bool) {
	node.edgeCount = edgeCount
	node.covered = covered
	heap.Fix(pq, node.index_min)
}

type Node struct {
	id, edgeCount, index_min, index_max int // edgeCount is used for sorting, index is used for quick access
	covered                             bool
}

type CoverCmd struct {
	pq_max *MaxPriorityQueue
	pq_min *MinPriorityQueue
	vc     []bool
	edges  [][]*Node
}

func (cmd *CoverCmd) Cover(node *Node) {
	if cmd.vc[node.id] {
		panic(fmt.Sprintf("%v already covered", node.id))
	}
	cmd.pq_max.update(node, -1, true)
	cmd.pq_min.update(node, -1, true)
	cmd.vc[node.id] = true

	// update neighbours
	for _, nn := range cmd.edges[node.id] {
		if !nn.covered {
			count := nn.edgeCount - 1
			cmd.pq_max.update(nn, count, nn.covered)
			cmd.pq_min.update(nn, count, nn.covered)
		}
	}
}

func (cmd *CoverCmd) Uncover(node *Node) {
	if !cmd.vc[node.id] {
		panic(fmt.Sprintf("%v not covered", node.id))
	}
	count := cmd.CountUncovered(node)
	cmd.pq_max.update(node, count, false)
	cmd.pq_min.update(node, count, false)
	cmd.vc[node.id] = false

	// update neighbours
	for _, nn := range cmd.edges[node.id] {
		if !nn.covered {
			count := nn.edgeCount + 1
			cmd.pq_max.update(nn, count, nn.covered)
			cmd.pq_min.update(nn, count, nn.covered)
		}
	}
}

func (cmd *CoverCmd) IsCovered(node *Node) bool {
	return cmd.vc[node.id]
}

func (cmd *CoverCmd) Uncovered(node *Node) []*Node {
	edges := slices.Clone(cmd.edges[node.id])
	return slices.DeleteFunc(edges, func(nn *Node) bool {
		return cmd.vc[nn.id]
	})
}

func (cmd *CoverCmd) Neighbours(node *Node) []*Node {
	return cmd.edges[node.id]
}

func (cmd *CoverCmd) UncoveredNeighbour(node *Node) *Node {
	for _, nn := range cmd.edges[node.id] {
		if !cmd.IsCovered(nn) {
			return nn
		}
	}
	return nil
}

func (cmd *CoverCmd) CountUncovered(node *Node) int {
	count := 0
	for _, nn := range cmd.edges[node.id] {
		if !cmd.vc[nn.id] {
			count++
		}
	}
	return count
}

type Graph struct {
	countNode, countEdges int
	nodes                 []*Node
	vc                    []bool
	edges                 [][]*Node
	pq_max                *MaxPriorityQueue
	pq_min                *MinPriorityQueue
}

func (g *Graph) AddNodeNoop(id int) *Node {
	if g.nodes[id] == nil {
		g.nodes[id] = &Node{
			id:        id,
			edgeCount: 0,
			index_min: -1,
			index_max: -1,
			covered:   false,
		}
		g.edges[id] = make([]*Node, 0, 1)
		g.pq_max.Push(g.nodes[id])
		g.pq_min.Push(g.nodes[id])
	}
	return g.nodes[id]
}

func (g *Graph) AddEdge(n1, n2 *Node) {
	g.edges[n1.id] = append(g.edges[n1.id], n2)
	count := n1.edgeCount + 1
	g.pq_max.update(n1, count, n1.covered)
	g.pq_min.update(n1, count, n1.covered)

	g.edges[n2.id] = append(g.edges[n2.id], n1)
	count = n2.edgeCount + 1
	g.pq_max.update(n2, count, n2.covered)
	g.pq_min.update(n2, count, n2.covered)
}

// improve: make vc_branch search the minimum cover instead of seraching multiple covers of certains sizes
func (g *Graph) VertexCover() []*Node {
	for k := 0; k <= g.countNode; k++ {

		cover := CoverCmd{
			pq_max: g.pq_max,
			pq_min: g.pq_min,
			vc:     g.vc,
			edges:  g.edges,
		}

		// save pq
		backup_min := slices.Clone(*g.pq_min)
		backup_max := slices.Clone(*g.pq_max)

		s := g.vc_branch(k, cover)
		logger.Printf("-> %v\n", s)

		if s != nil {
			return s
		}

		clear(g.vc)
		// restore pq
		g.pq_min = &backup_min
		for i, node := range backup_min {
			node.index_min = i
			//logger.Printf("reset %v.index = %v\n", node.id, node.index)
		}
		g.pq_max = &backup_max
		for i, node := range backup_max {
			node.index_max = i
			//logger.Printf("reset %v.index = %v\n", node.id, node.index)
		}
	}
	return nil
}

// improve: use fibonacci heap instead of binary heap, see https://github.com/Workiva/go-datastructures/tree/master
// improve: use lower and upper bounds to return earlier
func (g *Graph) vc_branch(k int, cover CoverCmd) []*Node {
	logger.Printf("vc_branch(k=%d)\n", k)
	logger.StepDown()
	defer logger.StepUp()

	if k < 0 {
		return nil
	}

	if g.pq_min.Len() > 0 {
		u := (*g.pq_min)[0]
		if !cover.IsCovered(u) {
			if v := cover.UncoveredNeighbour(u); v != nil && cover.CountUncovered(u) == 1 {
				logger.Printf("one degree rule - %v\n", u.id)
				cover.Cover(u)
				cover.Cover(v)
				if s := g.vc_branch(k-1, cover); s != nil {
					return append(s, v)
				}
				cover.Uncover(u)
				cover.Uncover(v)
			}
		}
	}

	var s []*Node

	if g.pq_max.Len() == 0 {
		return make([]*Node, 0)
	}
	// uncovered nodes appear in the pq before covered ones
	u := (*g.pq_max)[0]
	logger.Printf("u=%+v\n", u)

	// if an uncovered one appreas no more vertecies can be added
	if u.covered {
		logger.Printf("pq is empty\n")
		return make([]*Node, 0, g.countEdges/2)
	}

	if cover.CountUncovered(u) == 0 {
		logger.Printf("u.edgeCount == 0\n")
		cover.Cover(u)
		if s = g.vc_branch(k, cover); s != nil {
			return s
		}
		cover.Uncover(u)
		return nil
	}

	// branch node with most numbers of neighbours
	cover.Cover(u)
	if s = g.vc_branch(k-1, cover); s != nil {
		return append(s, u)
	}
	cover.Uncover(u)

	// branch all uncovered neighbours instead
	uncovered_neighbours := cover.Uncovered(u)
	for _, v := range uncovered_neighbours {
		cover.Cover(v)
	}
	if s = g.vc_branch(k-len(uncovered_neighbours), cover); s != nil {
		s = append(s, uncovered_neighbours...)
		return s
	}
	for _, v := range uncovered_neighbours {
		cover.Uncover(v)
	}

	return nil
}

func main() {
	g := parse(os.Stdin)
	nodes := g.VertexCover()
	for _, n := range nodes {
		fmt.Printf("%d\n", n.id+1)
	}
}

func parse(in io.Reader) *Graph {
	var countVertecies, countEdges int
	n, err := fmt.Fscanf(in, "%d %d\n", &countVertecies, &countEdges)
	if n != 2 {
		panic(err)
	}
	pq_max := make(MaxPriorityQueue, 0, countVertecies)
	pq_min := make(MinPriorityQueue, 0, countVertecies)
	g := &Graph{
		countNode:  countVertecies,
		countEdges: countEdges,
		nodes:      make([]*Node, countVertecies),
		edges:      make([][]*Node, countVertecies),
		vc:         make([]bool, countVertecies),
		pq_max:     &pq_max,
		pq_min:     &pq_min,
	}
	heap.Init(g.pq_max)
	heap.Init(g.pq_min)

	s := bufio.NewScanner(in)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		var src, trg int
		n, err := fmt.Sscanf(s.Text(), "%d %d", &src, &trg)
		if err == io.EOF {
			break
		}
		if n != 2 {
			panic(err)
		}
		if src == trg {
			panic(fmt.Sprintf("no self loops are allowed, got edge {%d, %d}\n", src, trg))
		}
		n1 := g.AddNodeNoop(src - 1)
		n2 := g.AddNodeNoop(trg - 1)
		g.AddEdge(n1, n2)
	}
	return g
}

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestVertexCover(t *testing.T) {
	problems := []string{
		//"vc1cg", // 240 probably lots of one degrees
		//"vc2ac", // 39
		"vc2ae", // 100
	}
	for _, problem := range problems {
		runTest(t, "./vc/in/"+problem+".in", "./vc/out/"+problem+".out")
		fmt.Printf("running test %s\n", problem)
	}

	RunVcBranch(t, 25)
}

func RunVcBranch(t *testing.T, k int) {
	problem := "vc3dm"
	in := "./vc/in/" + problem + ".in"
	res := "./vc/out/" + problem + ".out"

	fileIn, err := os.Open(in)
	if err != nil {
		panic(err)
	}
	defer fileIn.Close()

	fileRes, err := os.Open(res)
	if err != nil {
		panic(err)
	}
	defer fileRes.Close()

	var expected int
	fmt.Fscanf(fileRes, "%d", &expected)

	g := parse(fileIn)
	g.vc_branch(k, CoverCmd{
		pq_max: g.pq_max,
		pq_min: g.pq_min,
		vc:     g.vc,
		edges:  g.edges,
	})
}

func runTest(t *testing.T, in, res string) {
	fileIn, err := os.Open(in)
	if err != nil {
		panic(err)
	}
	defer fileIn.Close()

	fileRes, err := os.Open(res)
	if err != nil {
		panic(err)
	}
	defer fileRes.Close()

	var expected int
	fmt.Fscanf(fileRes, "%d", &expected)

	g := parse(fileIn)
	vc := g.VertexCover()
	//fmt.Printf("expected %d vertecies got %d\n", expected, len(vc))
	if expected != len(vc) {
		var cover strings.Builder
		cover.WriteString("[ ")
		for _, n := range vc {
			cover.WriteString(fmt.Sprintf("%+v ", n.id+1))
		}
		cover.WriteString("]")
		t.Fatalf("expected %d vertecies got %d: %v", expected, len(vc), cover.String())
	}
}

package recipe

import (
	"fmt"
	"os"
	"os/exec"
)

// Maximum number of concurrent dot processes
var maxConcurrentDot = 4
var dotSemaphore = make(chan struct{}, maxConcurrentDot)

// AddEdge adds an edge to the visualization
func AddEdge(from, to string) {
	visualMutex.Lock()
	defer visualMutex.Unlock()
	
	// Check if edge already exists
	for _, e := range currentEdges {
		if e.From == from && e.To == to {
			return // Edge already exists
		}
	}
	
	currentEdges = append(currentEdges, Edge{From: from, To: to})
}

// MarkVisited marks a node as visited in the visualization
func MarkVisited(node string, visited bool) {
	visualMutex.Lock()
	defer visualMutex.Unlock()
	
	currentVisited[node] = visited
}

// GetNextFrameID gets the next frame ID in a thread-safe way
func GetNextFrameID() int {
	frameMutex.Lock()
	defer frameMutex.Unlock()
	
	frameID := frameCount
	frameCount++
	return frameID
}

// ResetVisualization resets the visualization state
func ResetVisualization() {
	visualMutex.Lock()
	defer visualMutex.Unlock()
	
	currentEdges = []Edge{}
	currentVisited = make(map[string]bool)
	
	frameMutex.Lock()
	frameCount = 0
	frameMutex.Unlock()
}

// CaptureFrame captures the current state as a visualization frame
func CaptureFrame(elements map[string][][]string, basicElements map[string]bool) {
	if !VisualEnabled {
		return
	}
	
	visualMutex.Lock()
	
	// Skip if no edges (empty frame)
	if len(currentEdges) == 0 {
		visualMutex.Unlock()
		return
	}
	
	edges := make([]Edge, len(currentEdges))
	copy(edges, currentEdges)
	
	visited := make(map[string]bool)
	for k, v := range currentVisited {
		visited[k] = v
	}
	visualMutex.Unlock()
	
	// Get next frame ID
	frameID := GetNextFrameID()
	
	// Save the frame (non-blocking)
	go SaveDotFrame(edges, visited, basicElements, frameID)
}

// TraceLive creates a visualization of the DFS process
func TraceLive(path []string, elements map[string][][]string, basicElements map[string]bool) {
	if !VisualEnabled {
		return
	}
	
	// Reset visualization state
	ResetVisualization()
	
	// Create edges for the path
	for i := 0; i < len(path)-1; i++ {
		AddEdge(path[i], path[i+1])
	}
	
	// Capture the initial state
	CaptureFrame(elements, basicElements)
}

// SaveDotFrame saves a frame as a dot file and renders it to PNG
func SaveDotFrame(edges []Edge, visited map[string]bool, basic map[string]bool, frame int) {
	// Skip saving if no edges
	if len(edges) == 0 {
		return
	}
	
	// Acquire token for dot rendering (with non-blocking to prevent deadlock)
	select {
	case dotSemaphore <- struct{}{}:
		// Got the token, continue
		defer func() { <-dotSemaphore }()
	default:
		// Couldn't get token, skip rendering this frame
		return
	}
	
	// Create frames directory if it doesn't exist
	if _, err := os.Stat("frames"); os.IsNotExist(err) {
		os.Mkdir("frames", 0755)
	}
	
	filename := fmt.Sprintf("frames/step_%03d.dot", frame)
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	
	f.WriteString("digraph G {\n")
	f.WriteString(" node [shape=box, style=filled];\n")
	
	// Track all nodes from edges
	nodes := map[string]bool{}
	for _, e := range edges {
		nodes[e.From] = true
		nodes[e.To] = true
		// Use correct direction for the edge
		fmt.Fprintf(f, " \"%s\" -> \"%s\";\n", e.From, e.To)
	}
	
	// Style nodes (basic elements, visited nodes, etc.)
	for n := range nodes {
		color := "white"
		if basic[n] {
			color = "lightgreen" // Basic elements in green
		}
		if visited[n] {
			color = "lightskyblue" // Currently visited nodes in blue
		}
		
		// Add shape for target node (elements we're searching for) vs ingredient nodes
		shape := "box"
		style := "filled"
		
		fmt.Fprintf(f, " \"%s\" [shape=%s, style=%s, fillcolor=%s];\n", n, shape, style, color)
	}
	
	// Set layout direction 
	f.WriteString(" rankdir=TB;\n")
	f.WriteString("}\n")
	
	// Render PNG
	outPng := fmt.Sprintf("frames/step_%03d.png", frame)
	cmd := exec.Command("dot", "-Tpng", filename, "-o", outPng)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error rendering PNG for frame %d: %v\n", frame, err)
	}
}
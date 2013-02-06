// Copyright 2013, Friedrich Paetzke. All rights reserved.

// This package  provides wrapper functions for dot.
//
// Example:
/*
	dotter, err := godot.NewDotter(godot.OUT_PDF, godot.GRAPH_DIRECTED, "test.pdf")
	defer dotter.Close()
	if err != nil {
		panic(err)
	}

	dotter.SetLink("val", "val")
	dotter.SetLabel("val", "LABEL!!")

	dotter.SetLink("val", "val1")
	dotter.SetLink("val", "val1")
	dotter.SetLink("val", "val1")

	dotter.SetLabel("val3", "sLABEL!!")

	dotter.SetNodeShape("val3", godot.SHAPE_BOX)
	dotter.SetNodeShape("val34", godot.SHAPE_BOX)
*/
package godot

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type GraphType string
type NodeShape string
type Program string
type OutputType string

const (
	GRAPH_DIRECTED   GraphType = "digraph"
	GRAPH_UNDIRECTED GraphType = "graph"
)

const (
	SHAPE_BOX       NodeShape = "BOX"
	SHAPE_CIRCLE    NodeShape = "CIRCLE"
	SHAPE_FOLDER    NodeShape = "FOLDER"
	SHAPE_PLAINTEXT NodeShape = "PLAINTEXT"
	SHAPE_TRIANGLE  NodeShape = "TRIANGLE"
)

const (
	PROG_CIRCO Program = "circo"
	PROG_DOT   Program = "dot"
	PROG_FDP   Program = "fdp"
	PROG_NEATO Program = "neato"
	PROG_SFDP  Program = "sfdp"
	PROG_TWOPI Program = "twopi"
)

const (
	OUT_BMP OutputType = "bmp"
	OUT_DOT OutputType = "dot"
	OUT_JPG OutputType = "jpg"
	OUT_PDF OutputType = "pdf"
	OUT_PNG OutputType = "png"
	OUT_PS  OutputType = "ps"
	OUT_SVG OutputType = "svg"
)

type Dotter struct {
	instance   *exec.Cmd
	stdin      io.WriteCloser
	graphType  GraphType
	isStrict   bool
	isFirstCmd bool
}

func esc(node string) string {
	node = strings.Replace(node, ".", "DOT", -1)
	node = strings.Replace(node, "/", "SLASH", -1)
	node = strings.Replace(node, "-", "HYPHEN", -1)
	return node
}

func (dotter *Dotter) sendCmd(format string, args ...interface{}) error {
	if dotter.isFirstCmd {
		dotter.isFirstCmd = false
		if dotter.isStrict {
			dotter.sendCmd("strict")
		}
		dotter.sendCmd(string(dotter.graphType) + "{")
	}

	cmd := fmt.Sprintf(format, args...) + "\n"
	_, err := io.WriteString(dotter.stdin, cmd)
	return err
}

// Creates a New Dotter.
//
// Parameters:
//
// - isStrict: if true, multiple edges won't be displayed.
//
// - writeToFile: if true, output will be written to fname. Otherwise to stdout.
//
// - fname: filename. if fname equals "", dot will make up a filename - usally
// noname.dot.*
func NewDotterEx(oType OutputType, prog Program, gType GraphType,
	isStrict, writeToFile bool, fname string) (*Dotter, error) {

	dotPath, err := exec.LookPath(string(prog))
	if err != nil {
		panic(err)
	}

	otype := "-T" + string(oType)

	var cmd *exec.Cmd
	if writeToFile {
		ofile := "-O"
		if fname != "" {
			ofile = "-o" + fname
		}
		cmd = exec.Command(dotPath, otype, ofile)
	} else {
		cmd = exec.Command(dotPath, otype)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		go io.Copy(os.Stdin, stdout)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	dotter := &Dotter{}
	dotter.instance = cmd
	dotter.stdin = stdin
	dotter.graphType = gType
	dotter.isFirstCmd = true
	dotter.isStrict = isStrict

	return dotter, cmd.Start()
}

// Convenience Wrapper for NewDotterEx(). Makes proper assumptions.
// For more see NewDotterEx().
func NewDotter(oType OutputType, gType GraphType, fname string) (*Dotter, error) {
	return NewDotterEx(oType, PROG_DOT, gType, true, true, fname)
}

func (dotter *Dotter) Close() error {
	dotter.sendCmd("}")
	dotter.stdin.Close()
	return dotter.instance.Wait()
}

func (dotter *Dotter) SetLink(from, to string) error {
	link := "%s -- %s"
	if dotter.graphType == GRAPH_DIRECTED {
		link = "%s -> %s"
	}
	return dotter.sendCmd(link, esc(from), esc(to))
}

func (dotter *Dotter) SetLabel(node, label string) error {
	return dotter.sendCmd(`%s [label="%s"]`, esc(node), label)
}

func (dotter *Dotter) SetNodeShape(node string, shape NodeShape) error {
	return dotter.sendCmd(`%s [shape="%s"]`, esc(node), shape)
}

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/biogo/biogo/alphabet"
)

var matrices = []struct {
	file  string
	alpha alphabet.Alphabet
}{
	{"NUC.4", alphabet.DNA},
	{"NUC.4.4", alphabet.DNAredundant},

	{"DAYHOFF", alphabet.Protein},
	{"GONNET", alphabet.Protein},
	{"IDENTITY", alphabet.Protein},
	{"MATCH", alphabet.Protein},

	{"BLOSUM100", alphabet.Protein},
	{"BLOSUM30", alphabet.Protein},
	{"BLOSUM35", alphabet.Protein},
	{"BLOSUM40", alphabet.Protein},
	{"BLOSUM45", alphabet.Protein},
	{"BLOSUM50", alphabet.Protein},
	{"BLOSUM55", alphabet.Protein},
	{"BLOSUM60", alphabet.Protein},
	{"BLOSUM62", alphabet.Protein},
	{"BLOSUM65", alphabet.Protein},
	{"BLOSUM70", alphabet.Protein},
	{"BLOSUM75", alphabet.Protein},
	{"BLOSUM80", alphabet.Protein},
	{"BLOSUM85", alphabet.Protein},
	{"BLOSUM90", alphabet.Protein},
	{"BLOSUMN", alphabet.Protein},

	{"PAM10", alphabet.Protein},
	{"PAM100", alphabet.Protein},
	{"PAM110", alphabet.Protein},
	{"PAM120", alphabet.Protein},
	{"PAM120.cdi", alphabet.Protein},
	{"PAM130", alphabet.Protein},
	{"PAM140", alphabet.Protein},
	{"PAM150", alphabet.Protein},
	{"PAM160", alphabet.Protein},
	{"PAM160.cdi", alphabet.Protein},
	{"PAM170", alphabet.Protein},
	{"PAM180", alphabet.Protein},
	{"PAM190", alphabet.Protein},
	{"PAM20", alphabet.Protein},
	{"PAM200", alphabet.Protein},
	{"PAM200.cdi", alphabet.Protein},
	{"PAM210", alphabet.Protein},
	{"PAM220", alphabet.Protein},
	{"PAM230", alphabet.Protein},
	{"PAM240", alphabet.Protein},
	{"PAM250", alphabet.Protein},
	{"PAM250.cdi", alphabet.Protein},
	{"PAM260", alphabet.Protein},
	{"PAM270", alphabet.Protein},
	{"PAM280", alphabet.Protein},
	{"PAM290", alphabet.Protein},
	{"PAM30", alphabet.Protein},
	{"PAM300", alphabet.Protein},
	{"PAM310", alphabet.Protein},
	{"PAM320", alphabet.Protein},
	{"PAM330", alphabet.Protein},
	{"PAM340", alphabet.Protein},
	{"PAM350", alphabet.Protein},
	{"PAM360", alphabet.Protein},
	{"PAM370", alphabet.Protein},
	{"PAM380", alphabet.Protein},
	{"PAM390", alphabet.Protein},
	{"PAM40", alphabet.Protein},
	{"PAM400", alphabet.Protein},
	{"PAM40.cdi", alphabet.Protein},
	{"PAM410", alphabet.Protein},
	{"PAM420", alphabet.Protein},
	{"PAM430", alphabet.Protein},
	{"PAM440", alphabet.Protein},
	{"PAM450", alphabet.Protein},
	{"PAM460", alphabet.Protein},
	{"PAM470", alphabet.Protein},
	{"PAM480", alphabet.Protein},
	{"PAM490", alphabet.Protein},
	{"PAM50", alphabet.Protein},
	{"PAM500", alphabet.Protein},
	{"PAM60", alphabet.Protein},
	{"PAM70", alphabet.Protein},
	{"PAM80", alphabet.Protein},
	{"PAM80.cdi", alphabet.Protein},
	{"PAM90", alphabet.Protein},
}

func main() {
	fmt.Fprintln(os.Stdout, `// DO NOT EDIT. This file was autogenerated by make.go.

// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package matrix provides a variety of alignment scoring matrices for sequence alignment.
package matrix

// All alignment scoring matrices are organised to allow direct lookup using alphabets
// defined in biogo/alphabet. Gap penalties are set to zero for all matrices and the I/L
// single letter amino acid code, "J", is included but not defined for all the protein
// scoring matrices.
var (`)
	for i, m := range matrices {
		if i != 0 {
			fmt.Fprintln(os.Stdout)
		}
		err := genCode(os.Stdout, m.file, m.alpha)
		if err != nil {
			log.Fatalf("Failed to create matrix source for %s: %v", m, err)
		}
	}
	fmt.Println(")")
}

func genCode(w io.Writer, m string, a alphabet.Alphabet) error {
	f, err := os.Open(m)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	s := string(b)
	var (
		ind       = a.LetterIndex()
		ref       []string
		perm      []int
		mat       [][]int
		row       int
		lastBlank bool
	)
	for _, l := range strings.Split(s, "\n") {
		nsl := noSpace(l)
		switch {
		case len(l) == 0:
			if !lastBlank {
				lastBlank = true
				fmt.Fprintln(w, "\t//")
			}
		case l[0] == ' ':
			ref = strings.Fields(nsl)
			perm = make([]int, a.Len())
			for i, l := range ref {
				li := ind[l[0]]
				if li < 0 {
					continue
				}
				perm[li] = i
			}
			mat = make([][]int, a.Len())
			for j := range mat {
				mat[j] = make([]int, a.Len())
			}
			fallthrough
		case l[0] == '#':
			lastBlank = false
			fmt.Fprintf(w, "\t// %s\n", l)
		default:
			lastBlank = false
			fmt.Fprintf(w, "\t// %s\n", l)
			for col, f := range strings.Fields(nsl)[1:] {
				mat[ind[ref[row][0]]][ind[ref[col][0]]], err = strconv.Atoi(f)
				if err != nil {
					return err
				}
			}
			row++
		}
	}
	fmt.Fprintf(w, "\t%s = [][]int{\n\t\t/*       ", strings.Replace(m, ".", "_", -1))
	for j := range mat {
		fmt.Printf("%c ", toUpper(a.Letter(j)))
	}
	fmt.Fprintln(w, "*/")
	for i := range mat {
		fmt.Printf("\t\t/* %c */ {", toUpper(a.Letter(i)))
		for j, e := range mat[i] {
			fmt.Fprint(w, e)
			if j < len(mat[i])-1 {
				fmt.Print(", ")
			}
		}
		fmt.Fprintln(w, "},")
	}
	fmt.Fprintln(w, "\t}")
	return nil
}

func toUpper(l alphabet.Letter) alphabet.Letter {
	if l >= 'a' {
		return l &^ ' '
	}
	return l
}

func noSpace(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			if i > 0 && s[i-1] != ' ' {
				b = append(b, ' ')
			}
			continue
		}
		b = append(b, s[i])
	}
	return string(b)
}

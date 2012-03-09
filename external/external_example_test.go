// Package external allows uniform interaction with external tools.
package external

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"
	"strings"
)

func ExampleBuild_1() {
	// samtools sort [-n] [-m maxMem] <in.bam> <out.prefix>
	type SamToolsSort struct {
		Name      string
		Comment   string
		Cmd       string   `buildarg:"{{if .}}{{.}}{{else}}samtools{{end}}"` // samtools
		SubCmd    struct{} `buildarg:"sort"`                                 // sort
		SortNames bool     `buildarg:"{{if .}}-n{{end}}"`                    // [-n]
		MaxMem    int      `buildarg:"{{with .}}-m {{.}}{{end}}"`            // [-m maxMem]
		InFile    string   `buildarg:"\"{{.}}\""`                            // "<in.bam>"
		OutFile   string   `buildarg:"\"{{.}}\""`                            // "<out.prefix>"
		CommandBuilder
	}

	s := SamToolsSort{
		Name:      "Sort",
		SortNames: true,
		MaxMem:    1e8,
		InFile:    "infile",
		OutFile:   "outfile",
	}

	args, err := Build(s)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(strings.Join(args, " "))
	}
	// Output:
	// samtools sort -n -m 100000000 "infile" "outfile"
}

func ExampleBuild_2() {
	// samtools merge [-h inh.sam] [-n] <out.bam> <in1.bam> <in2.bam> [...]
	type SamToolsMerge struct {
		Name       string
		Comment    string
		Cmd        string   `buildarg:"{{if .}}{{.}}{{else}}samtools{{end}}"` // samtools
		SubCmd     struct{} `buildarg:"merge"`                                // merge
		HeaderFile string   `buildarg:"{{with .}}-h \"{{.}}\"{{end}}"`        // [-h inh.sam]
		SortNames  bool     `buildarg:"{{if .}}-n{{end}}"`                    // [-n]
		OutFile    string   `buildarg:"\"{{.}}\""`                            // "<out.bam>"
		InFiles    []string `buildarg:"{{quote . | join \" \"}}"`             // "<in.bam>"...
		CommandBuilder
	}

	s := &SamToolsMerge{
		Name:       "Merge",
		Cmd:        "samtools",
		HeaderFile: "header",
		InFiles:    []string{"infile1", "infile2"},
		OutFile:    "outfile",
	}

	args, err := Build(s)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(strings.Join(args, " "))
	}
	// Output:
	// samtools merge -h "header" "outfile" "infile1" "infile2"
}

func ExampleBuild_3() {
	// sed [-n] [-e <exp>]... [-f <file>]... [--follow-symlinks] [-i[suf]] [-l <len>] [--posix] [-r] [-s] [-s] <in>... > <out>
	type InPlace struct {
		Yes bool
		Suf string
	}
	type Sed struct {
		Name       string
		Comment    string
		Cmd        string   `buildarg:"{{if .}}{{.}}{{else}}sed{{end}}"`                   // sed
		Quiet      bool     `buildarg:"{{if .}}-n{{end}}"`                                 // [-n]
		Script     []string `buildarg:"{{mprintf \"-e '%v'\" . | join \" \"}}"`            // [-e '<exp>']...
		ScriptFile []string `buildarg:"{{mprintf \"-f %q\" . | join \" \"}}"`              // [-f "<file>"]...
		Follow     bool     `buildarg:"{{if .}}--follow-symlinks{{end}}"`                  // [--follow-symlinks]
		InPlace    InPlace  `buildarg:"{{if .Yes}}-i{{with .Suf}}\"{{.}}\"{{end}}{{end}}"` // [-i[suf]]
		WrapAt     int      `buildarg:"{{with .}}-l \"{{.}}\"{{end}}"`                     // [-l <len>]
		Posix      bool     `buildarg:"{{if .}}--posix{{end}}"`                            // [--posix]
		ExtendRE   bool     `buildarg:"{{if .}}-r{{end}}"`                                 // [-r]
		Separate   bool     `buildarg:"{{if .}}-s{{end}}"`                                 // [-s]
		Unbuffered bool     `buildarg:"{{if .}}-u{{end}}"`                                 // [-u]
		InFiles    []string `buildarg:"{{quote . | join \" \"}}"`                          // "<in>"...
		OutFile    string   `buildarg:"{{if .}}>\"{{.}}\"{{end}}"`                         // >"<out>"
		CommandBuilder
	}

	s := &Sed{
		Name:    "Sed",
		Cmd:     "sed",
		Script:  []string{`s/\<hi\>/lo/g`, `s/\<left\>/right/g`},
		InPlace: InPlace{true, "bottomright"},
		InFiles: []string{"infile"},
		OutFile: "outfile",
	}

	args, err := Build(s)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(strings.Join(args, " "))
	}
	// Output:
	// sed -e 's/\<hi\>/lo/g' -e 's/\<left\>/right/g' -i"bottomright" "infile" >"outfile"
}

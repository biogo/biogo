// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

// Package external allows uniform interaction with external tools based on a config struct.
package external

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"text/template"
)

// CommandBuilder is an interface that assembles a set of command line arguments, and creates
// an *exec.Cmd that can run the command. The method BuildCommand is responsible for handling 
// set up of redirections and parameter sanity checking if required. 
type CommandBuilder interface {
	BuildCommand() (*exec.Cmd, error)
}

// Mprintf applies Sprintf with the provided format to each element of slice. It returns an
// error if slice is not a slice or an array or a pointer to either of these types.
func Mprintf(format string, slice interface{}) (f []string, err error) {
	v := reflect.ValueOf(slice)
	if kind := v.Kind(); kind == reflect.Interface || kind == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		l := v.Len()
		f = make([]string, l)
		for i := 0; i < l; i++ {
			f[i] = fmt.Sprintf(format, v.Index(i).Interface())
		}
	default:
		return nil, errors.New("not a slice or array type")
	}

	return
}

// Quote wraps in quotes each string of a slice.
func Quote(s []string) (q []string) {
	q = make([]string, len(s))
	for i, u := range s {
		q[i] = fmt.Sprintf("%q", u)
	}

	return
}

// Join calls strings.Join with the parameter order reversed to allow use in a template pipeline.
func Join(sep string, a []string) string { return strings.Join(a, sep) }

// Build builds a set of command line args from cb, which must be a struct. cb's fields
// are inspected for struct tags "buildarg" key. The value for buildarg tag should be a valid
// text template.
// Template functions can be provided via funcs. Three convenience functions are provided:
//  quote is a template function that wraps elements of a slice of strings in quotes.
//  mprintf is a template function that applies fmt.Sprintf to each element of a slice.
//  join is a template function that calls strings.Join with parameter order reversed.
func Build(cb CommandBuilder, funcs ...template.FuncMap) (args []string, err error) {
	v := reflect.ValueOf(cb)
	if kind := v.Kind(); kind == reflect.Interface || kind == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}
	n := v.NumField()
	t := v.Type()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("buildarg")
		if tag != "" {
			tmpl := template.New(f.Name)
			tmpl.Funcs(template.FuncMap{
				"join":    Join,
				"quote":   Quote,
				"mprintf": Mprintf,
			})
			for _, f := range funcs {
				tmpl.Funcs(f)
			}
			if err != nil {
				return args, err
			}
			_, err = tmpl.Parse(tag)
			if err != nil {
				return args, err
			}
			b := &bytes.Buffer{}
			err = tmpl.Execute(b, v.Field(i).Interface())
			if err != nil {
				return args, err
			}
			if b.Len() > 0 {
				args = append(args, string(b.Bytes()))
			}
		}
	}

	return
}

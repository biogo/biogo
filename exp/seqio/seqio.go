// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Packages for reading and writing sequence files
package seqio

import "code.google.com/p/biogo/exp/seq"

// A SequenceAppender is a generic sequence type that can append elements.
type SequenceAppender interface {
	SetName(string)
	SetDescription(string)
	seq.Appender
	seq.Sequence
}

type Reader interface {
	Read() (seq.Sequence, error)
}

type Writer interface {
	Write(seq.Sequence) (int, error)
}

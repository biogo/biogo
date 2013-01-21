/* 
bíogo is a bioinformatics library for the Go language. It is a work in progress.

The Purpose of bíogo

bíogo stems from the need to address the size and structure of modern
genomic and metagenomic data sets. These properties enforce requirements on the
libraries and languages used for analysis:

	• speed - size of data sets
	• concurrency - problems often embarrassingly parallelisable

In addition to the computational burden of massive data set sizes in modern
genomics there is an increasing need for complex pipelines to resolve questions
in tightening problem space and also a developing need to be able to develop
new algorithms to allow novel approaches to interesting questions. These issues
suggest the need for a simplicity in syntax to facilitate:

	• ease of coding
	• checking for correctness in development and particularly in peer review

Related to the second issue is the reluctance of some researchers to release
code because of quality concerns
http://www.nature.com/news/2010/101013/full/467753a.html

The issue of code release is the first of the principles formalised in the
Science Code Manifesto http://sciencecodemanifesto.org/

 Code	All source code written specifically to process data for a published
	paper must be available to the reviewers and readers of the paper.

A language with a simple, yet expressive, syntax should facilitate development
of higher quality code and thus help reduce this barrier to research code
release.

Yet Another Bioinformatics Library

It seems that nearly every language has it own bioinformatics library, some of
which are very mature, for example BioPerl and BioPython. Why add another one?

The different libraries excel in different fields, acting as scripting glue for
applications in a pipeline (much of [1-3]) and interacting with external hosts
[1, 2, 4, 5], wrapping lower level high performance languages with more user
friendly syntax [1-4] or providing bioinformatics functions for high
performance languages [5, 6].

The intended niche for bíogo lies somewhere between the scripting libraries
and high performance language libraries in being easy to use for both small and
large projects while having reasonable performance with computationally
intensive tasks.

The intent is to reduce the level of investment required to develop new
research software for computationally intensive tasks.

 [1] BioPerl http://bioperl.org/
 	http://genome.cshlp.org/content/12/10/1611.full
 	http://www.springerlink.com/content/pp72033m171568p2

 [2] BioPython http://biopython.org/
	http://bioinformatics.oxfordjournals.org/content/25/11/1422

 [3] BioRuby http://bioruby.org/
 	http://bioinformatics.oxfordjournals.org/content/26/20/2617

 [4] PyCogent http://pycogent.sourceforge.net/
 	http://genomebiology.com/2007/8/8/R171

 [5] BioJava http://biojava.org/
	http://bioinformatics.oxfordjournals.org/content/24/18/2096

 [6] SeqAn http://www.seqan.de/
 	http://www.biomedcentral.com/1471-2105/9/11

Library Structure and Coding Style

The bíogo library structure is influenced both by the structure of BioPerl and
the Go core libraries.

The coding style should be aligned with normal Go idioms as represented in the
Go core libraries.

Position Numbering

Position numbering in the bíogo library conforms to the zero-based indexing
of Go and range indexing conforms to Go's half-open zero-based slice indexing.
This is at odds with the 'normal' inclusive indexing used by molecular
biologists. This choice was made to avoid inconsistent indexing spaces being
used — one-based inclusive for bíogo functions and methods and zero-based for
native Go slices and arrays — and so avoid errors that this would otherwise
facilitate.  Note that the GFF package does allow, and defaults to, one-based
inclusive indexing in its input and output of GFF files.

	EWD831 Why numbering should start at zero

	To denote the subsequence of natural numbers 2, 3, ..., 12 without the
	pernicious three dots, four conventions are open to us 

	a) 2 ≤ i< 13
	b) 1 < i≤ 12
	c) 2 ≤ i≤ 12
	d) 1 < i< 13

	Are there reasons to prefer one convention to the other? Yes, there are.
	The observation that conventions a) and b) have the advantage that the
	difference between the bounds as mentioned equals the length of the
	subsequence is valid. So is the observation that, as a consequence, in
	either convention two subsequences are adjacent means that the upper
	bound of the one equals the lower bound of the other. Valid as these
	observations are, they don't enable us to choose between a) and b); so
	let us start afresh.

	There is a smallest natural number. Exclusion of the lower bound —as in
	b) and d)— forces for a subsequence starting at the smallest natural
	number the lower bound as mentioned into the realm of the unnatural
	numbers. That is ugly, so for the lower bound we prefer the ≤ as in a)
	and c). Consider now the subsequences starting at the smallest natural
	number: inclusion of the upper bound would then force the latter to be
	unnatural by the time the sequence has shrunk to the empty one. That is
	ugly, so for the upper bound we prefer < as in a) and d). We conclude
	that convention a) is to be preferred.

	Remark  The programming language Mesa, developed at Xerox PARC, has
	special notations for intervals of integers in all four conventions.
	Extensive experience with Mesa has shown that the use of the other three
	conventions has been a constant source of clumsiness and mistakes, and
	on account of that experience Mesa programmers are now strongly advised
	not to use the latter three available features. I mention this
	experimental evidence —for what it is worth— because some people feel
	uncomfortable with conclusions that have not been confirmed in practice.
	(End of Remark.)

				*                * 
					*

	When dealing with a sequence of length N, the elements of which we wish
	to distinguish by subscript, the next vexing question is what subscript
	value to assign to its starting element. Adhering to convention a)
	yields, when starting with subscript 1, the subscript range 1 ≤  i <
	N+1; starting with 0, however, gives the nicer range 0 ≤   i <  N. So
	let us let our ordinals start at zero: an element's ordinal (subscript)
	equals the number of elements preceding it in the sequence. And the
	moral of the story is that we had better regard —after all those
	centuries!— zero as a most natural number.

	Remark  Many programming languages have been designed without due
	attention to this detail. In FORTRAN subscripts always start at 1; in
	ALGOL 60 and in PASCAL, convention c) has been adopted; the more recent
	SASL has fallen back on the FORTRAN convention: a sequence in SASL is at
	the same time a function on the positive integers. Pity! (End of
	Remark.)

				*                * 
					*

	The above has been triggered by a recent incident, when, in an emotional
	outburst, one of my mathematical colleagues at the University —not a
	computing scientist— accused a number of younger computing scientists of
	"pedantry" because —as they do by habit— they started numbering at zero.
	He took consciously adopting the most sensible convention as a
	provocation. (Also the "End of ..." convention is viewed of as
	provocative; but the convention is useful: I know of a student who
	almost failed at an examination by the tacit assumption that the
	questions ended at the bottom of the first page.) I think Antony Jay is
	right when he states: "In corporate religions as in others, the heretic
	must be cast out not because of the probability that he is wrong but
	because of the possibility that he is right."


	Plataanstraat 5		11 August 1982
	5671 AL NUENEN		prof.dr. Edsger W. Dijkstra
	The Netherlands		Burroughs Research Fellow


Quality Scores

Quality scores are supported for all sequence types, including protein. Phred
and Solexa scoring systems are able to be read from files, however internal
representation of quality scores is with Phred, so there will be precision loss
in conversion. A Solexa quality score type is provided for use where this will
be a problem.

Copyright ©2011-2012 The bíogo Authors except where otherwise noted. All rights
reserved. Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
*/
package biogo

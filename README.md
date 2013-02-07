![bíogo](http://biogo.googlecode.com/git/biogo.png)

#bíogo

##Installation

        $ go get code.google.com/p/biogo/...

##Overview

bíogo is a bioinformatics library for the Go language.

##The Purpose of bíogo

bíogo stems from the need to address the size and structure of modern genomic
and metagenomic data sets. These properties enforce requirements on the
libraries and languages used for analysis:

* speed - size of data sets
* concurrency - problems often embarrassingly parallelisable

In addition to the computational burden of massive data set sizes in modern
genomics there is an increasing need for complex pipelines to resolve questions
in tightening problem space and also a developing need to be able to develop
new algorithms to allow novel approaches to interesting questions. These issues
suggest the need for a simplicity in syntax to facilitate:

* ease of coding
* checking for correctness in development and particularly in peer review

Related to the second issue is the [reluctance of some researchers to release
code because of quality
concerns](http://www.nature.com/news/2010/101013/full/467753a.html "Publish
your computer code: it is good enough. Nature 2010.").

The issue of code release is the first of the principles formalised in the
[Science Code Manifesto](http://sciencecodemanifesto.org/).

    Code  All source code written specifically to process data for a published
          paper must be available to the reviewers and readers of the paper.

A language with a simple, yet expressive, syntax should facilitate development
of higher quality code and thus help reduce this barrier to research code
release.

##Yet Another Bioinformatics Library

It seems that nearly every language has it own bioinformatics library, some of
which are very mature, for example [BioPerl](http://bioperl.org) and
[BioPython](http://biopython.org). Why add another one?

The different libraries excel in different fields, acting as scripting glue for
applications in a pipeline (much of [[1], [2], [3]]) and interacting with external hosts
[[1], [2], [4], [5]], wrapping lower level high performance languages with more user
friendly syntax [[1], [2], [3], [4]] or providing bioinformatics functions for high
performance languages [[5], [6]].

The intended niche for bíogo lies somewhere between the scripting libraries and
high performance language libraries in being easy to use for both small and
large projects while having reasonable performance with computationally
intensive tasks.

The intent is to reduce the level of investment required to develop new
research software for computationally intensive tasks.

[1]: http://bioperl.org/ "BioPerl"
[2]: http://biopython.org/ "BioPython"
[3]: http://bioruby.org/ "BioRuby"
[4]: http://pycogent.sourceforge.net/ "PyCogent"
[5]: http://biojava.org/ "BioJava"
[6]: http://www.seqan.de/ "SeqAn"

1. BioPerl
    http://genome.cshlp.org/content/12/10/1611.full
    http://www.springerlink.com/content/pp72033m171568p2

2. BioPython
    http://bioinformatics.oxfordjournals.org/content/25/11/1422

3. BioRuby
    http://bioinformatics.oxfordjournals.org/content/26/20/2617

4. PyCogent
    http://genomebiology.com/2007/8/8/R171

5. BioJava
    http://bioinformatics.oxfordjournals.org/content/24/18/2096

6. SeqAn
    http://www.biomedcentral.com/1471-2105/9/11

##Library Structure and Coding Style

The bíogo library structure is influenced both by the Go core library.

The coding style should be aligned with normal Go idioms as represented in the
Go core libraries.

##Quality Scores

Quality scores are supported for all sequence types, including protein. Phred
and Solexa scoring systems are able to be read from files, however internal
representation of quality scores is with Phred, so there will be precision loss
in conversion. A Solexa quality score type is provided for use where this will
be a problem.

##Copyright and License

Copyright ©2011-2013 The bíogo Authors except where otherwise noted. All rights
reserved. Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.

The bíogo logo is derived from Bitstream Charter, Copyright ©1989-1992
Bitstream Inc., Cambridge, MA.

BITSTREAM CHARTER is a registered trademark of Bitstream Inc.

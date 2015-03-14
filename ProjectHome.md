![http://biogo.googlecode.com/git/biogo.png](http://biogo.googlecode.com/git/biogo.png)

## Installation ##

Assuming an existing [Go](http://golang.org) installation:

`$ go get code.google.com/p/biogo/...`

## Citing ##

If you use bíogo, please cite Kortschak and Adelson "bíogo: a simple high-performance bioinformatics toolkit for the Go language", doi:[10.1101/005033](http://biorxiv.org/content/early/2014/05/12/005033).

## Documentation ##

### Core packages ###

http://godoc.org/code.google.com/p/biogo/...

See sub-packages below.

### Mailing list ###

http://groups.google.com/group/biogo-user

## Overview ##

bíogo is a bioinformatics library for the Go language.

## The Purpose of bíogo ##

bíogo stems from the need to address the size and structure of modern genomic and metagenomic data sets. These properties enforce requirements on the libraries and languages used for analysis:

  * speed - size of data sets

  * concurrency - problems often embarrassingly parallelisable

In addition to the computational burden of massive data set sizes in modern genomics there is an increasing need for complex pipelines to resolve questions in tightening problem space and also a developing need to be able to develop new algorithms to allow novel approaches to interesting questions. These issues suggest the need for a simplicity in syntax to facilitate:

  * ease of coding

  * checking for correctness in development and particularly in peer review

These ideas are more fully discussed in [this paper](http://arxiv.org/abs/1210.0530).

Related to the second issue is the reluctance of some researchers to release code because of quality concerns ("[Publish your computer code: it is good enough. Nature 2010.](http://www.nature.com/news/2010/101013/full/467753a.html)").

The issue of code release is the first of the principles formalised in the [Science Code Manifesto](http://sciencecodemanifesto.org/).

<pre>
Code  All source code written specifically to process data for a published<br>
paper must be available to the reviewers and readers of the paper.<br>
</pre>

A language with a simple, yet expressive, syntax should facilitate development of higher quality code and thus help reduce this barrier to research code release. The [Go language design](http://talks.golang.org/2012/splash.article) satisfies these requirements.

If you use bíogo for work that you subsequently publish, please include a note in the paper linking to this site - and let us know.

## Articles ##

[bíogo: a simple high-performance bioinformatics toolkit for the Go language](http://biorxiv.org/content/early/2014/05/12/005033)

[Analysis of Illumina sequencing data using bíogo](http://talks.godoc.org/code.google.com/p/biogo.talks/illumination/illumina.article)

[Using and extending types in bíogo](http://talks.godoc.org/code.google.com/p/biogo.talks/types/types.article)

## Yet Another Bioinformatics Library ##

It seems that nearly every language has it own bioinformatics library, some of which are very mature, for example [BioPerl](http://bioperl.org) and [BioPython](http://biopython.org). Why add another one?

The different libraries excel in different fields, acting as scripting glue for applications in a pipeline (much of [[1](http://bioperl.org/)], [[2](http://biopython.org/)] and [[3](http://bioruby.org/)]) and interacting with external hosts[¹](http://bioperl.org/)<sup>, </sup>[²](http://biopython.org/)<sup>, </sup>[⁴](http://pycogent.sourceforge.net/)<sup>, </sup>[⁵](http://biojava.org/), wrapping lower level high performance languages with more user friendly syntax[¹](http://bioperl.org/)<sup>, </sup>[²](http://biopython.org/)<sup>, </sup>[³](http://bioruby.org/)<sup>, </sup>[⁴](http://pycogent.sourceforge.net/) or providing bioinformatics functions for high performance languages[⁵](http://biojava.org/)<sup>, </sup>[⁶](http://www.seqan.de/).

The intended niche for bíogo lies somewhere between the scripting libraries and high performance language libraries in being easy to use for both small and large projects while having reasonable performance with computationally intensive tasks.

The intent is to reduce the level of investment required to develop new research software for computationally intensive tasks.

  1. [BioPerl](http://bioperl.org/)
> > http://genome.cshlp.org/content/12/10/1611.full<br>
<blockquote><a href='http://www.springerlink.com/content/pp72033m171568p2'>http://www.springerlink.com/content/pp72033m171568p2</a>
</blockquote><ol><li><a href='http://biopython.org/'>BioPython</a>
<blockquote><a href='http://bioinformatics.oxfordjournals.org/content/25/11/1422'>http://bioinformatics.oxfordjournals.org/content/25/11/1422</a>
</blockquote></li><li><a href='http://bioruby.org/'>BioRuby</a>
<blockquote><a href='http://bioinformatics.oxfordjournals.org/content/26/20/2617'>http://bioinformatics.oxfordjournals.org/content/26/20/2617</a>
</blockquote></li><li><a href='http://pycogent.sourceforge.net/'>PyCogent</a>
<blockquote><a href='http://genomebiology.com/2007/8/8/R171'>http://genomebiology.com/2007/8/8/R171</a>
</blockquote></li><li><a href='http://biojava.org/'>BioJava</a>
<blockquote><a href='http://bioinformatics.oxfordjournals.org/content/24/18/2096'>http://bioinformatics.oxfordjournals.org/content/24/18/2096</a>
</blockquote></li><li><a href='http://www.seqan.de/'>SeqAn</a>
<blockquote><a href='http://www.biomedcentral.com/1471-2105/9/11'>http://www.biomedcentral.com/1471-2105/9/11</a></blockquote></li></ol>

<h3>Sub-packages</h3>

External resources:<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.ncbi/'>http://godoc.org/code.google.com/p/biogo.ncbi/</a>...<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.external/'>http://godoc.org/code.google.com/p/biogo.external/</a>

<a href='http://godoc.org/code.google.com/p/biogo.ragel/'>http://godoc.org/code.google.com/p/biogo.ragel/</a>

Graphics:<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.graphics/rings/'>http://godoc.org/code.google.com/p/biogo.graphics/rings/</a>

<a href='http://godoc.org/code.google.com/p/biogo.graphics/palette/'>http://godoc.org/code.google.com/p/biogo.graphics/palette/</a>...<br>
<br>
Stores:<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.store/llrb/'>http://godoc.org/code.google.com/p/biogo.store/llrb/</a>

<a href='http://godoc.org/code.google.com/p/biogo.store/interval/'>http://godoc.org/code.google.com/p/biogo.store/interval/</a>

<a href='http://godoc.org/code.google.com/p/biogo.store/step/'>http://godoc.org/code.google.com/p/biogo.store/step/</a>

<a href='http://godoc.org/code.google.com/p/biogo.store/kdtree/'>http://godoc.org/code.google.com/p/biogo.store/kdtree/</a>

Cgo-dependent packages:<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.matrix/'>http://godoc.org/code.google.com/p/biogo.matrix/</a> (package requires libcblas)<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.boom/'>http://godoc.org/code.google.com/p/biogo.boom/</a>

Examples:<br>
<br>
<a href='http://godoc.org/code.google.com/p/biogo.examples/'>http://godoc.org/code.google.com/p/biogo.examples/</a>

<h2>Quality Scores</h2>

Quality scores are supported for all sequence types, including protein. Phred and Solexa scoring systems are able to be read from files, however internal representation of quality scores is with Phred, so there will be precision loss in conversion. A Solexa quality score type is provided for use where this will be a problem.<br>
<br>
<h2>Contributing</h2>

If you are interested in contributing, please contact the <a href='http://groups.google.com/group/biogo-dev'>developer list</a>.<br>
<br>
<h2>Copyright and License</h2>

Copyright ©2011-2013 The bíogo Authors except where otherwise noted.<br>
<br>
See the <a href='http://code.google.com/p/biogo/source/browse/LICENSE'>LICENSE</a> file for license details.<br>
<br>
<br>
The bíogo logo is derived from Bitstream Charter, Copyright ©1989-1992 Bitstream Inc., Cambridge, MA.<br>
<br>
BITSTREAM CHARTER is a registered trademark of Bitstream Inc.
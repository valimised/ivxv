#! /bin/bash

DIR=$1
TARGET=`python3 -c "import sys; import os; sys.path.insert(0, os.path.abspath('..')); import documents; print(documents.get('${DIR}','document_target_name'));"`
LATEX=_build/latex
FULLTARGET=${LATEX}/${TARGET}
latexdiff --config VERBATIMENV=sphinxVerbatim _build/ivxv/Documentation/et/${DIR}/${FULLTARGET}.tex ${DIR}/${FULLTARGET}.tex > ${DIR}/${FULLTARGET}-diff.tex
cd ${DIR}/${LATEX}
pdflatex -interaction=nonstopmode ${TARGET}-diff.tex
pdflatex -interaction=nonstopmode ${TARGET}-diff.tex
cd ../../..
cp ${DIR}/${FULLTARGET}-diff.pdf _build/pdf/

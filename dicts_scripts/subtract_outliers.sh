#!/bin/bash

dos2unix $2

grep -F -x -v -f $2 $1/lexicon.tsv > lexicon.tmp && mv lexicon.tmp $1/lexicon.tsv

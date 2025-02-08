#!/bin/bash

cleanr=`wc -l ../../dicts/$1/clean_reverse.tsv | tr -c -d 0-9`
clean=`wc -l ../../dicts/$1/clean.tsv | tr -c -d 0-9`
dirty=`wc -l ../../dicts/$1/dirty.tsv | tr -c -d 0-9`

# Calculate percentages
percentage_clean=$((clean * 100 / dirty))
percentage_cleanr=$((cleanr * 100 / dirty))

# Print the results
echo "Coverage forward: $percentage_clean% for language $1"
echo "Coverage reverse: $percentage_cleanr% for language $1"

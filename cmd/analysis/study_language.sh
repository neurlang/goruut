#!/bin/bash

# Declare an array of phases
#phases_array=( 12 7 4 2 1 0 )
srcfile="../../dicts/$1/dirty.tsv"
file_size=$(cat $srcfile | wc -l)
phases_array=()
# more phases for bigger file. The phases are fibbonacci - 1
(( file_size > 150200 )) && phases_array=( "${phases_array[@]}" 20 )
(( file_size > 120200 )) && phases_array=( "${phases_array[@]}" 12 )
(( file_size >  90200 )) && phases_array=( "${phases_array[@]}"  7 )
(( file_size >  60200 )) && phases_array=( "${phases_array[@]}"  4 )
(( file_size >  30200 )) && phases_array=( "${phases_array[@]}"  2 )
(( file_size >    200 )) && phases_array=( "${phases_array[@]}"  1 )
phases_array=( "${phases_array[@]}" 0 )

for i in {0..100}; do 
	# Loop through each phase in the array
	for phase in "${phases_array[@]}"
	do
		echo "Phase $phase"
		./creator.sh $i $1 --randsubs $phase $2 $3 $4 $5 $6 ; 
		if [[ "$?" -eq 1 ]]; then
		    echo "Target met. Exiting."
		    exit
		fi
		./remover.sh $i $1 --randsubs $phase $2 $3 $4 $5 $6 ; 
		if [[ "$?" -eq 1 ]]; then
		    echo "Target met. Exiting."
		    exit
		fi
	done
done

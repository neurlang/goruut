#!/bin/bash

# Declare an array of phases
phases_array=( 12 7 4 2 1 0 )

for i in {0..100}; do 
	# Loop through each string in the array
	for phase in "${phases_array[@]}"
	do
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

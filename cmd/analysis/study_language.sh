#!/bin/bash

for i in {1..100}; do 
./creator.sh $i $1 $2 $3 $4 $5 $6 $7 $8 ; 
./remover.sh $i $1 $2 $3 $4 $5 $6 $7 $8 ; 
done

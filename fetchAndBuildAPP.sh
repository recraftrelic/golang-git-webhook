#!/bin/bash

eval "$(ssh-agent -s)"
ssh-add -D
ssh-add $6
cd $1
git checkout $2
git pull origin $2
eval $3
eval $4
eval $5
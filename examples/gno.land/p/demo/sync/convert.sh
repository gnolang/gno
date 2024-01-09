#!/bin/bash

for i in *.go; do g="$(echo $i | sed 's/.go$/.gno/')"; mv $i $g; done

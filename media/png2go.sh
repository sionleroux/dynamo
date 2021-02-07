#!/bin/bash

filename=$1

if [[ -z $1 ]]; then
    echo 'argument 1 must be input filename' >&2
    exit 1
fi

name=$(echo $filename | sed -e 's/.png//' -e 's/^./\U&/')
goname=${filename%%.png}.go

echo -ne "package media

// $name is generated from $filename
var $name []uint8 = []uint8{\n\t" > $goname

# skip the first metadata line
# convert colours to 0s & 1s
# replace newlines with commas
convert $filename txt:- \
    | tail -n +2 \
    | sed -e '/43523D/s/.*/0/' -e '/C7F0D8/s/.*/1/' \
    | tr '\n' ',' \
    | sed 's/,/, /g' \
    | sed 's/ $//' \
    >> $goname

echo -e '\n}' >> $goname

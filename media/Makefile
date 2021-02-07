pngfiles = $(wildcard *.png)
gofiles = $(pngfiles:%.png=%.go)

.PHONY: all

all: $(gofiles)

%.go: %.png
	./png2go.sh $<

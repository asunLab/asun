# Conformance runner: asun-c
# Builds: ./asun-conformance-c
# Run from this directory: it loads ../../cases.json and ../../encode-cases.json.

CC      ?= cc
CFLAGS  ?= -O2 -Wall -Wextra -Wno-unused-parameter -Wno-sign-compare -std=gnu11
INCLUDES := -I../../../asun-c/include

all: asun-conformance-c

asun-conformance-c: main.c ../../../asun-c/include/asun.h ../../../asun-c/include/asun_value.h
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ main.c -lm

run: asun-conformance-c
	./asun-conformance-c

clean:
	rm -f asun-conformance-c

.PHONY: all run clean

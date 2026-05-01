# Conformance runner: asun-cpp
# Builds: ./asun-conformance-cpp
# Run from this directory: it loads ../../cases.json and ../../encode-cases.json.

CXX      ?= c++
CXXFLAGS ?= -std=c++17 -O2 -Wall -Wextra -Wno-sign-compare -Wno-unused-parameter
INCLUDES := -I../../../asun-cpp/include

all: asun-conformance-cpp

asun-conformance-cpp: main.cpp ../../../asun-cpp/include/asun.hpp ../../../asun-cpp/include/asun_value.hpp
	$(CXX) $(CXXFLAGS) $(INCLUDES) -o $@ main.cpp

run: asun-conformance-cpp
	./asun-conformance-cpp

clean:
	rm -f asun-conformance-cpp

.PHONY: all run clean

#!/bin/bash
go mod download

# Sample commands
# # VideoProcessing
# go run cmd/benchmark/* -tasks 1 -distribution seq -workflow-type VideoProcessing -force-remote
# go run cmd/benchmark/* -tasks 1 -distribution seq -workflow-type VideoProcessing

# # ImageProcessing
# go run cmd/benchmark/* -tasks 1 -distribution seq -workflow-type ImageProcessing
# go run cmd/benchmark/* -tasks 1 -distribution seq -workflow-type ImageProcessing -use-mem
# go run cmd/benchmark/* -tasks 1 -distribution seq -workflow-type ImageProcessing -force-remote

# Benchmark
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type VideoProcessing -force-remote
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type VideoProcessing
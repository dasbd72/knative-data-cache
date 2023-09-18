#!/bin/bash
go mod download

# Sample commands
# # VideoProcessing
# go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 1 -distribution seq -workflow-type VideoProcessing -force-remote
# go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 1 -distribution seq -workflow-type VideoProcessing

# # ImageProcessing
# go run cmd/benchmark/* -bucket stress-benchmark -source larger_image -tasks 1 -distribution seq -workflow-type ImageProcessing
# go run cmd/benchmark/* -bucket stress-benchmark -source larger_image -tasks 1 -distribution seq -workflow-type ImageProcessing -use-mem
# go run cmd/benchmark/* -bucket stress-benchmark -source larger_image -tasks 1 -distribution seq -workflow-type ImageProcessing -force-remote

# Benchmark
go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 5 -distribution seq -workflow-type VideoProcessing -force-remote
go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 5 -distribution seq -workflow-type VideoProcessing
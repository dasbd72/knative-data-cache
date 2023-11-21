#!/bin/bash
go mod download

# Benchmark
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type ImageProcessing -force-remote -url-postfix default.192.168.100.0.sslip.io
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type ImageProcessing -url-postfix default.192.168.100.0.sslip.io
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type VideoProcessing -force-remote -url-postfix default.192.168.100.0.sslip.io
go run cmd/benchmark/* -tasks 5 -distribution seq -workflow-type VideoProcessing -url-postfix default.192.168.100.0.sslip.io
#!/bin/bash
go mod download
go run cmd/benchmark/* -ntasks 1000
go run cmd/benchmark/* -ntasks 1000 -force-remote
go run cmd/benchmark/* -ntasks 1000
go run cmd/benchmark/* -ntasks 1000 -force-remote
go run cmd/benchmark/* -ntasks 1000
go run cmd/benchmark/* -ntasks 1000 -force-remote
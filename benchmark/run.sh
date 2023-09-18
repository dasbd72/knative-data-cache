#!/bin/bash
go mod download
go run cmd/benchmark/* -tasks 5 -distribution seq 

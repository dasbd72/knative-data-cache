#!/bin/bash
go mod download
go run cmd/benchmark/* -tasks 10 -distribution seq -bucket video-processing -source original-video -workflow-type VideoProcessing

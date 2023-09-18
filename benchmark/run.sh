#!/bin/bash
go mod download

# VideoProcessing
go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 1 -distribution seq -workflow-type VideoProcessing -force-remote
go run cmd/benchmark/* -bucket video-processing -source original-video -tasks 1 -distribution seq -workflow-type VideoProcessing
package main

import (
	"fmt"
)

func function_chain(index int, bucket string, source string, forceRemote bool, useMem bool, workflowType string) FunctionChainResult {
	var res FunctionChainResult

	fmt.Println("function_chain", index, "start")

	switch workflowType {
	case "ImageProcessing":
		res = chain_image_processing(index, bucket, source, forceRemote, useMem)
	case "VideoProcessing":
		res = chain_video_processing(index, bucket, source, forceRemote, useMem)
	default:
		fmt.Println("Unknown workflow type.")
	}

	fmt.Println("function_chain", index, "end")

	return res
}

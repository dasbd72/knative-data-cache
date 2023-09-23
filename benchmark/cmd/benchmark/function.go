package main

import (
	"fmt"
)

func function_chain(index int, flags Flags) FunctionChainResult {
	var res FunctionChainResult

	fmt.Println("function_chain", index, "start")

	switch flags.WorkflowType {
	case "ImageProcessing":
		res = chain_image_processing(index, flags)
	case "VideoProcessing":
		res = chain_video_processing(index, flags)
	default:
		fmt.Println("Unknown workflow type.")
	}

	fmt.Println("function_chain", index, "end")

	return res
}

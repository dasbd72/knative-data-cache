package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func chain_image_processing(index int, flags Flags) FunctionChainResult {
	bucket := "images-processing"
	source := "larger-image"
	object_list := []string{"DSC08867.JPG", "DSC08868.JPG", "DSC08869.JPG", "DSC08870.JPG", "DSC08871.JPG"}

	intermediate := fmt.Sprintf("%s_%d-scaled", source, index)

	start := time.Now()
	is_result := function_image_scale(bucket, source, object_list, intermediate, flags.ForceRemote, flags.UseMem)
	ir_result := function_image_recognition(bucket, intermediate, object_list, flags.ForceRemote, flags.UseMem)
	duration := time.Since(start)

	return FunctionChainResult{int64(start.UnixMicro()), duration.Seconds(), is_result, ir_result, VideoSplitResult{}, VideoTranscodeResult{}, VideoMergeResult{}}
}

func function_image_scale(bucket string, source string, object_list []string, destination string, forceRemote bool, useMem bool) ImageScaleResult {
	var req_data ImageScaleRequest = ImageScaleRequest{
		Bucket:      bucket,
		Source:      source,
		ObjectList:  object_list,
		Destination: destination,
		ForceRemote: forceRemote,
	}
	var res_data ImageScaleResponse

	req := new(bytes.Buffer)
	err := json.NewEncoder(req).Encode(req_data)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	url := "http://image-scale.default.127.0.0.1.sslip.io"
	if !useMem {
		url = "http://image-scale-disk.default.127.0.0.1.sslip.io"
	}
	res, err := http.Post(url, "application/json", req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	duration := time.Since(start)

	p := make([]byte, 1024)
	n, _ := res.Body.Read(p)
	err = json.Unmarshal(p[:n], &res_data)
	if err != nil {
		fmt.Println(string(p[:n]))
		// panic(err)
	}

	return ImageScaleResult{duration.Seconds(), res_data}
}

func function_image_recognition(bucket string, source string, object_list []string, forceRemote bool, useMem bool) ImageRecognitionResult {
	var req_data ImageRecognitionRequest = ImageRecognitionRequest{
		Bucket:      bucket,
		Source:      source,
		ObjectList:  object_list,
		ForceRemote: forceRemote,
	}
	var res_data ImageRecognitionResponse

	req := new(bytes.Buffer)
	err := json.NewEncoder(req).Encode(req_data)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	url := "http://image-recognition.default.127.0.0.1.sslip.io"
	if !useMem {
		url = "http://image-recognition-disk.default.127.0.0.1.sslip.io"
	}
	res, err := http.Post(url, "application/json", req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	duration := time.Since(start)

	p := make([]byte, 1024)
	n, _ := res.Body.Read(p)
	err = json.Unmarshal(p[:n], &res_data)
	if err != nil {
		fmt.Println(string(p[:n]))
		// panic(err)
	}

	return ImageRecognitionResult{duration.Seconds(), res_data}
}

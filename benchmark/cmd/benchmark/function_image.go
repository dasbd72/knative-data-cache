package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func chain_image_processing(index int, bucket string, source string, forceRemote bool, useMem bool) FunctionChainResult {
	intermediate := fmt.Sprintf("%s_%d-scaled", source, index)

	start := time.Now()
	is_result := function_image_scale(bucket, source, intermediate, forceRemote, useMem)
	ir_result := function_image_recognition(intermediate, forceRemote, useMem)
	duration := time.Since(start)

	return FunctionChainResult{int64(start.UnixMicro()), duration.Seconds(), is_result, ir_result, VideoSplitResult{}, VideoTranscodeResult{}, VideoMergeResult{}}
}

func function_image_scale(bucket string, source string, destination string, forceRemote bool, useMem bool) ImageScaleResult {
	var req_data ImageScaleRequest
	var res_data ImageScaleResponse

	req_data.Bucket = bucket
	req_data.Source = source
	req_data.Destination = destination
	req_data.ForceRemote = forceRemote
	req_data.ForceBackup = false

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

func function_image_recognition(source string, forceRemote bool, useMem bool) ImageRecognitionResult {
	var req_data ImageRecognitionRequest
	var res_data ImageRecognitionResponse

	req_data.Bucket = "stress-benchmark"
	req_data.Source = source
	req_data.ForceRemote = forceRemote
	req_data.ForceBackup = false

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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func chain_video_processing(index int, bucket string, source string, forceRemote bool, useMem bool) FunctionChainResult {
	split_file_dst := fmt.Sprintf("%s_%d-splitted", source, index)
	merge_file_dst := fmt.Sprintf("%s_%d-transcoded", source, index)
	dst := fmt.Sprintf("%s_%d-merged", source, index)

	start := time.Now()
	vs_result := function_video_split(bucket, source, split_file_dst, forceRemote, useMem)

	var transcode_result [5]VideoTranscodeResult
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			var splited_file_path = split_file_dst + "/seg" + strconv.Itoa(i+1) + "_sample.mp4"
			transcode_result[i] = function_video_transcode(bucket, splited_file_path, merge_file_dst, forceRemote, useMem)
		}(i)
	}
	wg.Wait()

	vt_result := VideoTranscodeResult{}
	vt_result.Duration = 0.0
	vt_result.Response = VideoTranscodeResponse{
		ForceRemote:       forceRemote,
		CodeDuration:      0.0,
		DownloadDuration:  0.0,
		TranscodeDuration: 0.0,
		UploadDuration:    0.0,
	}
	for i := 0; i < 5; i++ {
		vt_result.Duration += transcode_result[i].Duration
		vt_result.Response.CodeDuration += transcode_result[i].Response.CodeDuration
		vt_result.Response.DownloadDuration += transcode_result[i].Response.DownloadDuration
		vt_result.Response.TranscodeDuration += transcode_result[i].Response.TranscodeDuration
		vt_result.Response.UploadDuration += transcode_result[i].Response.UploadDuration
	}
	vt_result.Duration /= 5
	vt_result.Response.CodeDuration /= 5
	vt_result.Response.DownloadDuration /= 5
	vt_result.Response.TranscodeDuration /= 5
	vt_result.Response.UploadDuration /= 5

	vm_result := function_video_merge(bucket, merge_file_dst, dst, forceRemote, useMem)
	duration := time.Since(start)

	return FunctionChainResult{int64(start.UnixMicro()), duration.Seconds(), ImageScaleResult{}, ImageRecognitionResult{}, vs_result, vt_result, vm_result}
}

func function_video_split(bucket string, source string, destination string, forceRemote bool, useMem bool) VideoSplitResult {
	var req_data VideoSplitRequest
	var res_data VideoSplitResponse

	req_data.Bucket = bucket
	req_data.Source = source
	req_data.Destination = destination
	req_data.ForceRemote = forceRemote

	req := new(bytes.Buffer)
	err := json.NewEncoder(req).Encode(req_data)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	url := "http://video-split.default.127.0.0.1.sslip.io"
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

	return VideoSplitResult{duration.Seconds(), res_data}
}

func function_video_transcode(bucket string, source string, destination string, forceRemote bool, useMem bool) VideoTranscodeResult {
	var req_data VideoTranscodeRequest
	var res_data VideoTranscodeResponse

	req_data.Bucket = bucket
	req_data.Source = source
	req_data.Destination = destination
	req_data.ForceRemote = forceRemote

	req := new(bytes.Buffer)
	err := json.NewEncoder(req).Encode(req_data)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	url := "http://video-transcode.default.127.0.0.1.sslip.io"
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

	return VideoTranscodeResult{duration.Seconds(), res_data}
}

func function_video_merge(bucket string, source string, destination string, forceRemote bool, useMem bool) VideoMergeResult {
	var req_data VideoMergeRequest
	var res_data VideoMergeResponse

	req_data.Bucket = bucket
	req_data.Source = source
	req_data.Destination = destination
	req_data.ForceRemote = forceRemote

	req := new(bytes.Buffer)
	err := json.NewEncoder(req).Encode(req_data)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	url := "http://video-merge.default.127.0.0.1.sslip.io"
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

	return VideoMergeResult{duration.Seconds(), res_data}
}

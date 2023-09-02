package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ImageScaleRequest struct {
	Bucket      string `json:"bucket"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ForceRemote bool   `json:"force_remote"`
	ForceBackup bool   `json:"force_backup"`
}

type ImageScaleResponse struct {
	ForceRemote      bool    `json:"force_remote"`
	CodeDuration     float64 `json:"code_duration"`
	DownloadDuration float64 `json:"download_duration"`
	ScaleDuration    float64 `json:"scale_duration"`
	UploadDuration   float64 `json:"upload_duration"`
}

type ImageRecognitionRequest struct {
	Bucket      string `json:"bucket"`
	Source      string `json:"source"`
	ForceRemote bool   `json:"force_remote"`
	ForceBackup bool   `json:"force_backup"`
}

type ImageRecognitionResponse struct {
	Predictions       []string `json:"predictions"`
	ForceRemote       bool     `json:"force_remote"`
	ShortResult       bool     `json:"short_result"`
	CodeDuration      float64  `json:"code_duration"`
	DownloadDuration  float64  `json:"download_duration"`
	InferenceDuration float64  `json:"inference_duration"`
}

type ImageScaleResult struct {
	Duration float64            `json:"duration"`
	Response ImageScaleResponse `json:"response"`
}

type ImageRecognitionResult struct {
	Duration float64                  `json:"duration"`
	Response ImageRecognitionResponse `json:"response"`
}

type FunctionChainResult struct {
	StartTs  int64                  `json:"start_ts"`
	Duration float64                `json:"duration"`
	IsResult ImageScaleResult       `json:"is_result"`
	IrResult ImageRecognitionResult `json:"ir_result"`
}

func function_chain(index int, forceRemote bool) FunctionChainResult {
	fmt.Println("function_chain", index, "start")

	// ==================== function ====================
	source := "larger_image"
	intermediate := fmt.Sprintf("larger_image_%d-scaled", index)
	start := time.Now()
	is_result := function_image_scale(source, intermediate, forceRemote)
	// time.Sleep(10 * time.Second)
	ir_result := function_image_recognition(intermediate, forceRemote)
	duration := time.Since(start)

	fmt.Println("function_chain", index, "end")

	return FunctionChainResult{int64(start.UnixMicro()), duration.Seconds(), is_result, ir_result}
}

func function_image_scale(source string, destination string, forceRemote bool) ImageScaleResult {
	var req_data ImageScaleRequest
	var res_data ImageScaleResponse

	req_data.Bucket = "stress-benchmark"
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
	res, err := http.Post("http://image-scale.default.127.0.0.1.sslip.io", "application/json", req)
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

func function_image_recognition(source string, forceRemote bool) ImageRecognitionResult {
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
	res, err := http.Post("http://image-recognition.default.127.0.0.1.sslip.io", "application/json", req)
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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
	ForceRemote          bool    `json:"force_remote"`
	ForceBackup          bool    `json:"force_backup"`
	CodeDuration         float64 `json:"code_duration"`
	DownloadDuration     float64 `json:"download_duration"`
	ScaleDuration        float64 `json:"scale_duration"`
	UploadDuration       float64 `json:"upload_duration"`
	DownloadPostDuration float64 `json:"download_post_duration"`
	UploadPostDuration   float64 `json:"upload_post_duration"`
	BackupPostDuration   float64 `json:"backup_post_duration"`
}

type ImageRecognitionRequest struct {
	Bucket      string `json:"bucket"`
	Source      string `json:"source"`
	ForceRemote bool   `json:"force_remote"`
	ForceBackup bool   `json:"force_backup"`
}

type ImageRecognitionResponse struct {
	Predictions          []string `json:"predictions"`
	ForceRemote          bool     `json:"force_remote"`
	ForceBackup          bool     `json:"force_backup"`
	ShortResult          bool     `json:"short_result"`
	CodeDuration         float64  `json:"code_duration"`
	DownloadDuration     float64  `json:"download_duration"`
	InferenceDuration    float64  `json:"inference_duration"`
	DownloadPostDuration float64  `json:"download_post_duration"`
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

func function_chain(wg *sync.WaitGroup, index int, forceRemote bool) FunctionChainResult {
	defer wg.Done()

	fmt.Println("function_chain", index, "start")

	// ==================== function ====================
	source := fmt.Sprintf("larger_image_%d", index)
	start := time.Now()
	is_result := function_image_scale(source, forceRemote)
	ir_result := function_image_recognition(source+"-scaled", forceRemote)
	duration := time.Since(start)

	fmt.Println("function_chain", index, "end")

	return FunctionChainResult{int64(start.UnixMicro()), duration.Seconds(), is_result, ir_result}
}

func function_image_scale(source string, forceRemote bool) ImageScaleResult {
	var req_data ImageScaleRequest
	var res_data ImageScaleResponse

	req_data.Bucket = "stress-benchmark"
	req_data.Source = source
	req_data.Destination = source + "-scaled"
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

	err = json.NewDecoder(res.Body).Decode(&res_data)
	if err != nil {
		panic(err)
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

	err = json.NewDecoder(res.Body).Decode(&res_data)
	if err != nil {
		panic(err)
	}

	return ImageRecognitionResult{duration.Seconds(), res_data}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	ntasks      int
	rate        float64
	forceRemote bool
)

func init() {
	flag.IntVar(&ntasks, "ntasks", 10, "number of tasks")
	flag.Float64Var(&rate, "rate", 0.5, "rate of poisson process")
	flag.BoolVar(&forceRemote, "force-remote", false, "force remote")

	flag.Parse()
}

func main() {
	// ==================== log ====================
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// create logs directory
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	// open log file
	f, err := os.OpenFile(fmt.Sprintf("logs/%d.log", time.Now().Unix()), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// set log output to file
	log.SetOutput(f)

	// ==================== warm up ====================
	warmup()

	// ==================== benchmark ====================
	benchmark(ntasks, rate, forceRemote)
}

func warmup() {
	fmt.Println("warming up")

	wg := new(sync.WaitGroup)
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go function_chain(wg, i, forceRemote)
	}
	wg.Wait()

	fmt.Println("warm up done")
}

func benchmark(ntasks int, rate float64, forceRemote bool) {
	wg := new(sync.WaitGroup)
	wg.Add(ntasks)
	results := make([]FunctionChainResult, ntasks)

	start := time.Now()
	for i := 0; i < ntasks; i++ {
		// poisson process interval
		x := -math.Log(1.0-rand.Float64()) / rate
		// function invocation
		go func(i int) {
			results[i] = function_chain(wg, i, forceRemote)
		}(i)
		// sleep
		time.Sleep(time.Duration(x) * time.Second)
	}
	wg.Wait()
	duration := time.Since(start)

	// ==================== log ====================
	var benchmark_result struct {
		Duration      float64 `json:"duration"`
		TotalDuration float64 `json:"total_duration"`

		AverageDuration float64 `json:"average_duration"`

		TotalIsDuration         float64 `json:"total_is_duration"`
		TotalIsCodeDuration     float64 `json:"total_is_code_duration"`
		TotalIsDownloadDuration float64 `json:"total_is_download_duration"`
		TotalIsUploadDuration   float64 `json:"total_is_upload_duration"`

		AverageIsDuration         float64 `json:"average_is_duration"`
		AverageIsCodeDuration     float64 `json:"average_is_code_duration"`
		AverageIsDownloadDuration float64 `json:"average_is_download_duration"`
		AverageIsUploadDuration   float64 `json:"average_is_upload_duration"`

		TotalIrDuration         float64 `json:"total_ir_duration"`
		TotalIrCodeDuration     float64 `json:"total_ir_code_duration"`
		TotalIrDownloadDuration float64 `json:"total_ir_download_duration"`

		AverageIrDuration         float64 `json:"average_ir_duration"`
		AverageIrCodeDuration     float64 `json:"average_ir_code_duration"`
		AverageIrDownloadDuration float64 `json:"average_ir_download_duration"`
	}

	// full log
	p, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println(string(p))

	// ==================== result ====================
	benchmark_result.Duration = duration.Seconds()

	for _, result := range results {
		benchmark_result.TotalDuration += result.Duration

		benchmark_result.TotalIsDuration += result.IsResult.Duration
		benchmark_result.TotalIsCodeDuration += result.IsResult.Response.CodeDuration
		benchmark_result.TotalIsDownloadDuration += result.IsResult.Response.DownloadDuration
		benchmark_result.TotalIsUploadDuration += result.IsResult.Response.UploadDuration

		benchmark_result.TotalIrDuration += result.IrResult.Duration
		benchmark_result.TotalIrCodeDuration += result.IrResult.Response.CodeDuration
		benchmark_result.TotalIrDownloadDuration += result.IrResult.Response.DownloadDuration
	}
	benchmark_result.AverageDuration = benchmark_result.TotalDuration / float64(ntasks)

	benchmark_result.AverageIsDuration = benchmark_result.TotalIsDuration / float64(ntasks)
	benchmark_result.AverageIsCodeDuration = benchmark_result.TotalIsCodeDuration / float64(ntasks)
	benchmark_result.AverageIsDownloadDuration = benchmark_result.TotalIsDownloadDuration / float64(ntasks)
	benchmark_result.AverageIsUploadDuration = benchmark_result.TotalIsUploadDuration / float64(ntasks)

	benchmark_result.AverageIrDuration = benchmark_result.TotalIrDuration / float64(ntasks)
	benchmark_result.AverageIrCodeDuration = benchmark_result.TotalIrCodeDuration / float64(ntasks)
	benchmark_result.AverageIrDownloadDuration = benchmark_result.TotalIrDownloadDuration / float64(ntasks)

	p, err = json.MarshalIndent(benchmark_result, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println(string(p))
}

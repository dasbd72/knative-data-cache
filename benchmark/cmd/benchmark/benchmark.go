package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

var (
	flags struct {
		concurrency int
		tasks       int
		rate        float64
		forceRemote bool
		warmup      bool
	}
)

func init() {
	flag.IntVar(&flags.concurrency, "concurrency", 2147483647, "number of concurrent tasks")
	flag.IntVar(&flags.tasks, "tasks", 10, "number of tasks")
	flag.Float64Var(&flags.rate, "rate", 0.5, "rate of poisson process")
	flag.BoolVar(&flags.forceRemote, "force-remote", false, "force remote")
	flag.BoolVar(&flags.warmup, "warmup", false, "warmup")

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
	if flags.warmup {
		warmup()
	}

	// ==================== benchmark ====================
	benchmark(flags.concurrency, flags.tasks, flags.rate, flags.forceRemote)
}

func warmup() {
	fmt.Println("warming up")

	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			function_chain(i, true)
		}(i)
	}
	wg.Wait()

	fmt.Println("warm up done")
}

func benchmark(concurrency int, tasks int, rate float64, forceRemote bool) {
	fmt.Println("benchmarking")

	// ==================== benchmark ====================
	ctx := context.TODO()
	sem_cc := semaphore.NewWeighted(int64(concurrency))

	results := make([]FunctionChainResult, tasks)

	start := time.Now()
	for i := 0; i < tasks; i++ {
		// poisson process interval
		x := -math.Log(1.0-rand.Float64()) / rate

		// wait for concurrency control
		if err := sem_cc.Acquire(ctx, 1); err != nil {
			panic(err)
		}

		// function invocation
		go func(i int) {
			defer sem_cc.Release(1)

			result := function_chain(i, forceRemote)
			results[i] = result

			// logging
			p, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				panic(
					fmt.Sprintf(
						"index: %d, error: %s",
						i,
						err.Error(),
					),
				)
			}
			log.Println("index:", i, "\n", string(p))
		}(i)
		// sleep
		time.Sleep(time.Duration(x) * time.Second)
	}
	sem_cc.Acquire(ctx, int64(concurrency))
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
	benchmark_result.AverageDuration = benchmark_result.TotalDuration / float64(tasks)

	benchmark_result.AverageIsDuration = benchmark_result.TotalIsDuration / float64(tasks)
	benchmark_result.AverageIsCodeDuration = benchmark_result.TotalIsCodeDuration / float64(tasks)
	benchmark_result.AverageIsDownloadDuration = benchmark_result.TotalIsDownloadDuration / float64(tasks)
	benchmark_result.AverageIsUploadDuration = benchmark_result.TotalIsUploadDuration / float64(tasks)

	benchmark_result.AverageIrDuration = benchmark_result.TotalIrDuration / float64(tasks)
	benchmark_result.AverageIrCodeDuration = benchmark_result.TotalIrCodeDuration / float64(tasks)
	benchmark_result.AverageIrDownloadDuration = benchmark_result.TotalIrDownloadDuration / float64(tasks)

	p, err := json.MarshalIndent(benchmark_result, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println(string(p))
}

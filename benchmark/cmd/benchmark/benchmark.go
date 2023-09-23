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
	flags Flags
)

func init() {
	flag.IntVar(&flags.Concurrency, "concurrency", 2147483647, "number of concurrent tasks")
	flag.IntVar(&flags.Tasks, "tasks", 5, "number of tasks")
	flag.StringVar(&flags.Distribution, "distribution", "poisson", "distribution of tasks: [poisson, burst, seq|sequential]")
	flag.Float64Var(&flags.Rate, "rate", 0.5, "rate of poisson process")
	flag.BoolVar(&flags.ForceRemote, "force-remote", false, "force remote")
	flag.BoolVar(&flags.Warmup, "warmup", false, "warmup")
	flag.BoolVar(&flags.UseMem, "use-mem", false, "use memory if true, else use disk")
	flag.StringVar(&flags.WorkflowType, "workflow-type", "ImageProcessing", "workflow type")

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

	// ==================== flags ====================
	p, err := json.MarshalIndent(flags, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println("flags:\n", string(p))

	// ==================== warm up ====================
	if flags.Warmup {
		warmup(flags)
	}

	// ==================== benchmark ====================
	benchmark(flags)
}

func warmup(flags Flags) {
	fmt.Println("warming up")

	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			function_chain(i, flags)
		}(i)
	}
	wg.Wait()

	fmt.Println("warm up done")
}

func benchmark(flags Flags) {
	fmt.Println("benchmarking")

	// ==================== benchmark ====================
	results := make([]FunctionChainResult, flags.Tasks)

	invoke := func(i int) {
		result := function_chain(i, flags)
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
	}

	start := time.Now()

	switch flags.Distribution {
	case "poisson":
		ctx := context.TODO()
		sem_cc := semaphore.NewWeighted(int64(flags.Concurrency))

		for i := 0; i < flags.Tasks; i++ {
			// poisson process interval
			x := -math.Log(1.0-rand.Float64()) / flags.Rate

			// wait for concurrency control
			if err := sem_cc.Acquire(ctx, 1); err != nil {
				panic(err)
			}

			// function invocation
			go func(i int) {
				defer sem_cc.Release(1)
				invoke(i)
			}(i)

			// sleep
			time.Sleep(time.Duration(x) * time.Second)
		}
		sem_cc.Acquire(ctx, int64(flags.Concurrency))

	case "burst":
		for i := 0; i < flags.Tasks; i++ {
			go invoke(i)
		}
	case "seq":
		fallthrough
	case "sequential":
		for i := 0; i < flags.Tasks; i++ {
			invoke(i)
		}
	default:
		panic("invalid distribution")
	}

	duration := time.Since(start)
	// ==================== benchmark done ====================

	var benchmark_result BenchmarkResult

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

		benchmark_result.TotalVsDuration += result.VsResult.Duration
		benchmark_result.TotalVsCodeDuration += result.VsResult.Response.CodeDuration
		benchmark_result.TotalVsDownloadDuration += result.VsResult.Response.DownloadDuration
		benchmark_result.TotalVsUploadDuration += result.VsResult.Response.UploadDuration

		benchmark_result.TotalVtDuration += result.VtResult.Duration
		benchmark_result.TotalVtCodeDuration += result.VtResult.Response.CodeDuration
		benchmark_result.TotalVtDownloadDuration += result.VtResult.Response.DownloadDuration
		benchmark_result.TotalVtUploadDuration += result.VtResult.Response.UploadDuration

		benchmark_result.TotalVmDuration += result.VmResult.Duration
		benchmark_result.TotalVmCodeDuration += result.VmResult.Response.CodeDuration
		benchmark_result.TotalVmDownloadDuration += result.VmResult.Response.DownloadDuration
		benchmark_result.TotalVmUploadDuration += result.VmResult.Response.UploadDuration
	}
	benchmark_result.AverageDuration = benchmark_result.TotalDuration / float64(flags.Tasks)

	benchmark_result.AverageIsDuration = benchmark_result.TotalIsDuration / float64(flags.Tasks)
	benchmark_result.AverageIsCodeDuration = benchmark_result.TotalIsCodeDuration / float64(flags.Tasks)
	benchmark_result.AverageIsDownloadDuration = benchmark_result.TotalIsDownloadDuration / float64(flags.Tasks)
	benchmark_result.AverageIsUploadDuration = benchmark_result.TotalIsUploadDuration / float64(flags.Tasks)

	benchmark_result.AverageIrDuration = benchmark_result.TotalIrDuration / float64(flags.Tasks)
	benchmark_result.AverageIrCodeDuration = benchmark_result.TotalIrCodeDuration / float64(flags.Tasks)
	benchmark_result.AverageIrDownloadDuration = benchmark_result.TotalIrDownloadDuration / float64(flags.Tasks)

	benchmark_result.AverageVsDuration = benchmark_result.TotalVsDuration / float64(flags.Tasks)
	benchmark_result.AverageVsCodeDuration = benchmark_result.TotalVsCodeDuration / float64(flags.Tasks)
	benchmark_result.AverageVsDownloadDuration = benchmark_result.TotalVsDownloadDuration / float64(flags.Tasks)
	benchmark_result.AverageVsUploadDuration = benchmark_result.TotalVsUploadDuration / float64(flags.Tasks)

	benchmark_result.AverageVtDuration = benchmark_result.TotalVtDuration / float64(flags.Tasks)
	benchmark_result.AverageVtCodeDuration = benchmark_result.TotalVtCodeDuration / float64(flags.Tasks)
	benchmark_result.AverageVtDownloadDuration = benchmark_result.TotalVtDownloadDuration / float64(flags.Tasks)
	benchmark_result.AverageVtUploadDuration = benchmark_result.TotalVtUploadDuration / float64(flags.Tasks)

	benchmark_result.AverageVmDuration = benchmark_result.TotalVmDuration / float64(flags.Tasks)
	benchmark_result.AverageVmCodeDuration = benchmark_result.TotalVmCodeDuration / float64(flags.Tasks)
	benchmark_result.AverageVmDownloadDuration = benchmark_result.TotalVmDownloadDuration / float64(flags.Tasks)
	benchmark_result.AverageVmUploadDuration = benchmark_result.TotalVmUploadDuration / float64(flags.Tasks)

	p, err := json.MarshalIndent(benchmark_result, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println(string(p))
}

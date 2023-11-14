package main

type Flags struct {
	UrlPostfix   string  `json:"url_postfix"`
	Concurrency  int     `json:"concurrency"`
	Tasks        int     `json:"tasks"`
	Distribution string  `json:"distribution"`
	Rate         float64 `json:"rate"`
	ForceRemote  bool    `json:"force_remote"`
	Warmup       bool    `json:"warmup"`
	WorkflowType string  `json:"workflow_type"`
}

type BenchmarkResult struct {
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

	TotalVsDuration         float64 `json:"total_vs_duration"`
	TotalVsCodeDuration     float64 `json:"total_vs_code_duration"`
	TotalVsDownloadDuration float64 `json:"total_vs_download_duration"`
	TotalVsUploadDuration   float64 `json:"total_vs_upload_duration"`

	AverageVsDuration         float64 `json:"average_vs_duration"`
	AverageVsCodeDuration     float64 `json:"average_vs_code_duration"`
	AverageVsDownloadDuration float64 `json:"average_vs_download_duration"`
	AverageVsUploadDuration   float64 `json:"average_vs_upload_duration"`

	TotalVtDuration         float64 `json:"total_vt_duration"`
	TotalVtCodeDuration     float64 `json:"total_vt_code_duration"`
	TotalVtDownloadDuration float64 `json:"total_vt_download_duration"`
	TotalVtUploadDuration   float64 `json:"total_vt_upload_duration"`

	AverageVtDuration         float64 `json:"average_vt_duration"`
	AverageVtCodeDuration     float64 `json:"average_vt_code_duration"`
	AverageVtDownloadDuration float64 `json:"average_vt_download_duration"`
	AverageVtUploadDuration   float64 `json:"average_vt_upload_duration"`

	TotalVmDuration         float64 `json:"total_vm_duration"`
	TotalVmCodeDuration     float64 `json:"total_vm_code_duration"`
	TotalVmDownloadDuration float64 `json:"total_vm_download_duration"`
	TotalVmUploadDuration   float64 `json:"total_vm_upload_duration"`

	AverageVmDuration         float64 `json:"average_vm_duration"`
	AverageVmCodeDuration     float64 `json:"average_vm_code_duration"`
	AverageVmDownloadDuration float64 `json:"average_vm_download_duration"`
	AverageVmUploadDuration   float64 `json:"average_vm_upload_duration"`
}

// -----------------------------------------------
type ImageScaleRequest struct {
	Bucket      string   `json:"bucket"`
	Source      string   `json:"source"`
	ObjectList  []string `json:"object_list"`
	Destination string   `json:"destination"`
	ForceRemote bool     `json:"force_remote"`
}

type ImageScaleResponse struct {
	ForceRemote      bool    `json:"force_remote"`
	CodeDuration     float64 `json:"code_duration"`
	DownloadDuration float64 `json:"download_duration"`
	ScaleDuration    float64 `json:"scale_duration"`
	UploadDuration   float64 `json:"upload_duration"`
}

type ImageRecognitionRequest struct {
	Bucket      string   `json:"bucket"`
	Source      string   `json:"source"`
	ObjectList  []string `json:"object_list"`
	ForceRemote bool     `json:"force_remote"`
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

// ---------------------------------------
type VideoSplitRequest struct {
	Bucket      string   `json:"bucket"`
	Source      string   `json:"source"`
	ObjectList  []string `json:"object_list"`
	Destination string   `json:"destination"`
	ForceRemote bool     `json:"force_remote"`
}
type VideoSplitResponse struct {
	ForceRemote      bool    `json:"force_remote"`
	CodeDuration     float64 `json:"code_duration"`
	DownloadDuration float64 `json:"download_duration"`
	SplitDuration    float64 `json:"split_duration"`
	UploadDuration   float64 `json:"upload_duration"`
}
type VideoSplitResult struct {
	Duration float64            `json:"duration"`
	Response VideoSplitResponse `json:"response"`
}

type VideoTranscodeRequest struct {
	Bucket      string   `json:"bucket"`
	Source      string   `json:"source"`
	ObjectList  []string `json:"object_list"`
	Destination string   `json:"destination"`
	ForceRemote bool     `json:"force_remote"`
}
type VideoTranscodeResponse struct {
	ForceRemote       bool    `json:"force_remote"`
	CodeDuration      float64 `json:"code_duration"`
	DownloadDuration  float64 `json:"download_duration"`
	TranscodeDuration float64 `json:"transcode_duration"`
	UploadDuration    float64 `json:"upload_duration"`
}
type VideoTranscodeResult struct {
	Duration float64                `json:"duration"`
	Response VideoTranscodeResponse `json:"response"`
}

type VideoMergeRequest struct {
	Bucket      string   `json:"bucket"`
	Source      string   `json:"source"`
	ObjectList  []string `json:"object_list"`
	Destination string   `json:"destination"`
	ForceRemote bool     `json:"force_remote"`
}
type VideoMergeResponse struct {
	ForceRemote      bool    `json:"force_remote"`
	CodeDuration     float64 `json:"code_duration"`
	DownloadDuration float64 `json:"download_duration"`
	MergeDuration    float64 `json:"merge_duration"`
	UploadDuration   float64 `json:"upload_duration"`
}
type VideoMergeResult struct {
	Duration float64            `json:"duration"`
	Response VideoMergeResponse `json:"response"`
}

// ---------------------------------------------
type FunctionChainResult struct {
	StartTs  int64                  `json:"start_ts"`
	Duration float64                `json:"duration"`
	IsResult ImageScaleResult       `json:"is_result"`
	IrResult ImageRecognitionResult `json:"ir_result"`
	VsResult VideoSplitResult       `json:"vs_result"`
	VtResult VideoTranscodeResult   `json:"vt_result"`
	VmResult VideoMergeResult       `json:"vm_result"`
}

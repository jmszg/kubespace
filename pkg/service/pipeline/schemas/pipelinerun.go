package schemas

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"time"
)

// JobCallbackParams 任务执行完成回调参数
type JobCallbackParams struct {
	JobId  uint   `json:"job_id"`
	Status string `json:"status"`
}

// AddReleaseVersionParams 发布阶段执行时添加版本
type AddReleaseVersionParams struct {
	WorkspaceId uint   `json:"workspace_id"`
	JobId       uint   `json:"job_id"`
	Version     string `json:"version"`
}

// PipelineStageCancelParams 流水线构建阶段取消
type PipelineStageCancelParams struct {
	StageRunId uint `json:"stage_run_id"`
}

// PipelineStageReexecParams 流水线构建阶段重新执行
type PipelineStageReexecParams struct {
	StageRunId uint `json:"stage_run_id"`
}

// PipelineBuildParams 流水线构建参数
type PipelineBuildParams struct {
	PipelineId    uint                         `json:"pipeline_id"`
	CodeBranch    *PipelineBuildCodeBranch     `json:"code_branch"`
	CustomSources []*PipelineBuildCustomSource `json:"custom_sources"`
	//Params        map[string]interface{}       `json:"params"`
}

// PipelineBuildCodeBranch 流水线代码空间构建参数
type PipelineBuildCodeBranch struct {
	Branch     string    `json:"branch"`
	CommitId   string    `json:"commit_id"`
	Author     string    `json:"author"`
	Message    string    `json:"message"`
	CommitTime time.Time `json:"commit_time"`
}

// PipelineBuildCustomSource 流水线自定义空间构建参数，当前流水线绑定的每个源
type PipelineBuildCustomSource struct {
	WorkspaceId         uint   `json:"workspace_id"`
	WorkspaceName       string `json:"workspace_name"`
	PipelineId          uint   `json:"pipeline_id"`
	PipelineName        string `json:"pipeline_name"`
	BuildReleaseVersion string `json:"build_release_version"`
	BuildId             uint   `json:"build_id"`
	BuildNumber         uint   `json:"build_number"`
	BuildOperator       string `json:"build_operator"`
	CodeAuthor          string `json:"code_author"`
	CodeBranch          string `json:"code_branch"`
	CodeComment         string `json:"code_comment"`
	CodeCommit          string `json:"code_commit"`
	CodeCommitTime      string `json:"code_commit_time"`
	IsBuild             bool   `json:"is_build" default:"true"`
}

type PipelineParams struct {
	ID          uint                  `json:"id"`
	WorkspaceId uint                  `json:"workspace_id"`
	Name        string                `json:"name"`
	Sources     types.PipelineSources `json:"sources"`
	Triggers    []*PipelineTrigger    `json:"triggers"`
	Stages      []*PipelineStage      `json:"stages"`
}

type PipelineTrigger struct {
	Id uint `json:"id"`
	// 触发类型，
	Type string `json:"type"`
	// 定时配置
	Cron string `json:"cron,omitempty"`
}

type PipelineStage struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	TriggerMode  string                 `json:"trigger_mode"`
	CustomParams map[string]interface{} `json:"custom_params"`
	Jobs         types.PipelineJobs     `json:"jobs"`
}

package adminreq

type TaskIDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type TaskListQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type TaskCreateRequest struct {
	Name              string  `json:"name" binding:"required,min=1,max=128"`
	Spec              string  `json:"spec" binding:"required,min=1,max=64"`
	Command           string  `json:"command" binding:"required,min=1,max=1024"`
	TimeoutSec        *int    `json:"timeout_sec" binding:"omitempty,min=0,max=3600"`
	ConcurrencyPolicy *string `json:"concurrency_policy" binding:"omitempty,oneof=forbid allow"`
	Enabled           *bool   `json:"enabled"`
}

type TaskUpdateRequest struct {
	Name              *string `json:"name" binding:"omitempty,min=1,max=128"`
	Spec              *string `json:"spec" binding:"omitempty,min=1,max=64"`
	Command           *string `json:"command" binding:"omitempty,min=1,max=1024"`
	TimeoutSec        *int    `json:"timeout_sec" binding:"omitempty,min=0,max=3600"`
	ConcurrencyPolicy *string `json:"concurrency_policy" binding:"omitempty,oneof=forbid allow"`
	Enabled           *bool   `json:"enabled"`
}

type TaskToggleRequest struct {
	Enabled bool `json:"enabled"`
}

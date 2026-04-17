package adminreq

type QueueTaskListQuery struct {
	Queue    string `form:"queue"`
	State    string `form:"state"` // retry | archived
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type QueueTaskActionURI struct {
	Queue  string `uri:"queue" binding:"required,min=1,max=64"`
	TaskID string `uri:"task_id" binding:"required,min=1,max=128"`
}

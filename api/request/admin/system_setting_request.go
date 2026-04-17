package adminreq

type SystemSettingIDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type SystemSettingListQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Key       string `form:"key"`
	GroupName string `form:"group_name"`
}

type SystemSettingHistoryQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type SystemSettingCreateRequest struct {
	Key       string `json:"key" binding:"required,min=1,max=128"`
	Value     string `json:"value" binding:"required"`
	ValueType string `json:"value_type" binding:"omitempty,oneof=string int bool json"`
	GroupName string `json:"group_name" binding:"omitempty,max=64"`
	Remark    string `json:"remark" binding:"omitempty,max=255"`
}

type SystemSettingUpdateRequest struct {
	Value     *string `json:"value"`
	ValueType *string `json:"value_type" binding:"omitempty,oneof=string int bool json"`
	GroupName *string `json:"group_name" binding:"omitempty,max=64"`
	Remark    *string `json:"remark" binding:"omitempty,max=255"`
}

type SystemSettingRollbackRequest struct {
	HistoryID int64  `json:"history_id" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"omitempty,max=255"`
}

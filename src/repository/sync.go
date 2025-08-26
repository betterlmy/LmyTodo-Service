package repository

// ===== 数据同步相关类型 =====

// TodoSyncItem TODO同步项
type TodoSyncItem struct {
	ID          int      `json:"id,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Completed   bool     `json:"completed"`
	Priority    int      `json:"priority"`
	DueDate     *string  `json:"due_date,omitempty"`
	Tags        []string `json:"tags"`
	CategoryID  *int     `json:"category_id,omitempty"`
	Reminder    *string  `json:"reminder,omitempty"`
	IsDeleted   bool     `json:"is_deleted"`
	SyncVersion int64    `json:"sync_version"`
	UpdatedAt   string   `json:"updated_at"`
}

// CategorySyncItem 分类同步项
type CategorySyncItem struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	IsDeleted   bool   `json:"is_deleted"`
	SyncVersion int64  `json:"sync_version"`
	UpdatedAt   string `json:"updated_at"`
}

// UserSettingsSyncItem 用户设置同步项
type UserSettingsSyncItem struct {
	Theme            string `json:"theme"`
	NotificationTime string `json:"notification_time"`
	Language         string `json:"language"`
	TimeZone         string `json:"timezone"`
	SyncVersion      int64  `json:"sync_version"`
	UpdatedAt        string `json:"updated_at"`
}

// SyncResult 同步结果
type SyncResult struct {
	Type        string `json:"type"`
	LocalID     int    `json:"local_id,omitempty"`
	ServerID    int    `json:"server_id,omitempty"`
	Action      string `json:"action"`
	Message     string `json:"message,omitempty"`
	SyncVersion int64  `json:"sync_version,omitempty"`
}

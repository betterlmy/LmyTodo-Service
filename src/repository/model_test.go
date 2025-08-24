package repository

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPriorityEnum(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
		{PriorityUrgent, "urgent"},
	}

	for _, test := range tests {
		if test.priority.String() != test.expected {
			t.Errorf("Priority.String() = %v, want %v", test.priority.String(), test.expected)
		}
	}
}

func TestStringSliceJSON(t *testing.T) {
	tags := StringSlice{"工作", "重要", "紧急"}

	// 测试序列化
	jsonData, err := json.Marshal(tags)
	if err != nil {
		t.Fatalf("Failed to marshal StringSlice: %v", err)
	}

	expected := `["工作","重要","紧急"]`
	if string(jsonData) != expected {
		t.Errorf("JSON marshal = %v, want %v", string(jsonData), expected)
	}

	// 测试反序列化
	var newTags StringSlice
	err = json.Unmarshal(jsonData, &newTags)
	if err != nil {
		t.Fatalf("Failed to unmarshal StringSlice: %v", err)
	}

	if len(newTags) != len(tags) {
		t.Errorf("Unmarshaled length = %v, want %v", len(newTags), len(tags))
	}

	for i, tag := range tags {
		if newTags[i] != tag {
			t.Errorf("Unmarshaled tag[%d] = %v, want %v", i, newTags[i], tag)
		}
	}
}

func TestTodoModelExtended(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	reminder := now.Add(1 * time.Hour)

	todo := Todo{
		ID:          1,
		UserID:      1,
		Title:       "测试任务",
		Description: "这是一个测试任务",
		Completed:   false,
		Priority:    PriorityHigh,
		DueDate:     &dueDate,
		Tags:        StringSlice{"工作", "重要"},
		CategoryID:  func() *int { id := 1; return &id }(),
		Reminder:    &reminder,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
		SyncVersion: now.UnixMilli(),
	}

	// 验证基本字段
	if todo.Title != "测试任务" {
		t.Errorf("Todo.Title = %v, want %v", todo.Title, "测试任务")
	}

	if todo.Priority != PriorityHigh {
		t.Errorf("Todo.Priority = %v, want %v", todo.Priority, PriorityHigh)
	}

	if len(todo.Tags) != 2 {
		t.Errorf("Todo.Tags length = %v, want %v", len(todo.Tags), 2)
	}

	if todo.Tags[0] != "工作" || todo.Tags[1] != "重要" {
		t.Errorf("Todo.Tags = %v, want %v", todo.Tags, []string{"工作", "重要"})
	}

	if todo.CategoryID == nil || *todo.CategoryID != 1 {
		t.Errorf("Todo.CategoryID = %v, want %v", todo.CategoryID, 1)
	}
}

func TestCategoryModel(t *testing.T) {
	category := Category{
		ID:        1,
		UserID:    1,
		Name:      "工作",
		Color:     "#FF5722",
		Icon:      "work",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}

	if category.Name != "工作" {
		t.Errorf("Category.Name = %v, want %v", category.Name, "工作")
	}

	if category.Color != "#FF5722" {
		t.Errorf("Category.Color = %v, want %v", category.Color, "#FF5722")
	}

	if category.Icon != "work" {
		t.Errorf("Category.Icon = %v, want %v", category.Icon, "work")
	}
}

func TestUserSettingsModel(t *testing.T) {
	settings := UserSettings{
		UserID:           1,
		Theme:            "dark",
		NotificationTime: "09:00:00",
		Language:         "zh-CN",
		TimeZone:         "Asia/Shanghai",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if settings.Theme != "dark" {
		t.Errorf("UserSettings.Theme = %v, want %v", settings.Theme, "dark")
	}

	if settings.Language != "zh-CN" {
		t.Errorf("UserSettings.Language = %v, want %v", settings.Language, "zh-CN")
	}

	if settings.TimeZone != "Asia/Shanghai" {
		t.Errorf("UserSettings.TimeZone = %v, want %v", settings.TimeZone, "Asia/Shanghai")
	}
}

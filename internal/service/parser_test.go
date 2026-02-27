package service

import (
	"Ralf/internal/domen"
	"reflect"
	"testing"
)

func TestGetNewTask(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    domen.Task
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Get one Task",
			path: "tasks.txt",
			want: domen.Task{
				Num:           1,
				Description:   "первая задача.",
				ImportantInfo: "первый важный момент.",
				ExpectResult:  "первый ожидаемый результат.",
				TestsValue:    "первые входные данные.",
				FuncSignature: "первая сигнатура функции.",
				Status:        domen.StatusNew,
			},
			wantErr: false,
		},
		{
			name:    "no task",
			path:    "no_tasks.txt",
			want:    domen.Task{},
			wantErr: true,
		},
		{
			name:    "invalid task",
			path:    "invalid_tasks.txt",
			want:    domen.Task{},
			wantErr: true,
		},
		{
			name:    "no tasks file",
			path:    "no_tasks_file.txt",
			want:    domen.Task{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNewTask(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNewTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNewTask() got = %v, want %v", got, tt.want)
			}
		})
	}
}

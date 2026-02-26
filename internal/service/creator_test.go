package service

import (
	"Ralf/domen"
	"testing"
)

func Test_updateStatusInLine(t *testing.T) {
	type args struct {
		line      string
		newStatus domen.TaskStatus
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "edit_new_status_to_run",
			args: args{line: "статус выполнения:new", newStatus: domen.StatusRun},
			want: "статус выполнения:run",
		},
		{
			name: "edit_new_status_to_error",
			args: args{line: "статус выполнения:new", newStatus: domen.StatusError},
			want: "статус выполнения:error",
		},
		{
			name: "edit_new_status_to_ok",
			args: args{line: "статус выполнения:new", newStatus: domen.StatusOK},
			want: "статус выполнения:ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateStatusInLine(tt.args.line, tt.args.newStatus); got != tt.want {
				t.Errorf("updateStatusInLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	type args struct {
		filePath  string
		taskNum   int
		newStatus domen.TaskStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "no file",
			args: args{
				filePath:  "no_file.txt",
				taskNum:   1,
				newStatus: domen.StatusRun,
			},
			wantErr: true,
		},
		{
			name: "no task",
			args: args{
				filePath:  "no_task.txt",
				taskNum:   2,
				newStatus: domen.StatusRun,
			},
			wantErr: true,
		},
		{
			name: "one task run",
			args: args{
				filePath:  "one_task.txt",
				taskNum:   1,
				newStatus: domen.StatusRun,
			},
			wantErr: false,
		},
		{
			name: "one task new",
			args: args{
				filePath:  "one_task.txt",
				taskNum:   1,
				newStatus: domen.StatusNew,
			},
			wantErr: false,
		},
		{
			name: "3 task run",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusRun,
			},
			wantErr: false,
		},
		{
			name: "3 task new for run",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusNew,
			},
			wantErr: false,
		},
		{
			name: "3 task ok",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusOK,
			},
			wantErr: false,
		},
		{
			name: "3 task new for ok",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusNew,
			},
			wantErr: false,
		},
		{
			name: "3 task error",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusError,
			},
			wantErr: false,
		},
		{
			name: "3 task new for error",
			args: args{
				filePath:  "three_task.txt",
				taskNum:   3,
				newStatus: domen.StatusNew,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateTaskStatus(tt.args.filePath, tt.args.taskNum, tt.args.newStatus); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTaskStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

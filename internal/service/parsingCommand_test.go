package service

import (
	"Ralf/domen"
	"reflect"
	"testing"
)

func Test_mapRussianType(t *testing.T) {
	type args struct {
		typStr string
	}
	tests := []struct {
		name  string
		args  args
		want  domen.CommandType
		want1 bool
	}{
		// TODO: Add test cases.
		{
			name: "no command",
			args: args{
				"нету команды",
			},
			want1: false,
		},
		{
			name: "create",
			args: args{
				"создание",
			},
			want:  domen.CmdCreate,
			want1: true,
		},
		{
			name: "delete",
			args: args{
				"удаление",
			},
			want:  domen.CmdDelete,
			want1: true,
		},
		{
			name: "edit",
			args: args{
				"внесение изменений",
			},
			want:  domen.CmdEdit,
			want1: true,
		},
		{
			name: "add new line",
			args: args{
				"добавление строк",
			},
			want:  domen.CmdAddLines,
			want1: true,
		},
		{
			name: "copy",
			args: args{
				"копирование",
			},
			want:  domen.CmdCopy,
			want1: true,
		},
		{
			name: "move",
			args: args{
				"перемещение",
			},
			want:  domen.CmdMove,
			want1: true,
		},
		{
			name: "read",
			args: args{
				"чтение",
			},
			want:  domen.CmdRead,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mapRussianType(tt.args.typStr)
			if got != tt.want {
				t.Errorf("mapRussianType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("mapRussianType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_parseLinesMap(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want map[int]string
	}{
		// TODO: Add test cases.
		{
			name: "add one line",
			args: args{"1:\"new text\""},
			want: map[int]string{1: "new text"},
		},
		{
			name: "add two lines",
			args: args{"1:\"new text\", 2:\"2 text\""},
			want: map[int]string{1: "new text", 2: "2 text"},
		},
		{
			name: "add 10 lines",
			args: args{
				"1:\"1 text\"," +
					" 2:\"2 text\"," +
					" 3:\"3 text\"," +
					" 4:\"4 text\"," +
					" 5:\"5 text\"," +
					" 6:\"6 text\"," +
					" 7:\"7 text\"," +
					" 8:\"8 text\"," +
					" 9:\"9 text\"," +
					" 10:\"10 text\""},
			want: map[int]string{
				1:  "1 text",
				2:  "2 text",
				3:  "3 text",
				4:  "4 text",
				5:  "5 text",
				6:  "6 text",
				7:  "7 text",
				8:  "8 text",
				9:  "9 text",
				10: "10 text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLinesMap(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLinesMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

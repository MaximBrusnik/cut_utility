package main

import (
	"testing"
)

func TestParseFieldRanges(t *testing.T) {
	tests := []struct {
		name    string
		fields  string
		want    []FieldRange
		wantErr bool
	}{
		{
			name:   "отдельные поля",
			fields: "1,3,5",
			want: []FieldRange{
				{Start: 1, End: 1},
				{Start: 3, End: 3},
				{Start: 5, End: 5},
			},
			wantErr: false,
		},
		{
			name:   "диапазоны",
			fields: "1-3,5-7",
			want: []FieldRange{
				{Start: 1, End: 3},
				{Start: 5, End: 7},
			},
			wantErr: false,
		},
		{
			name:   "смешанные поля и диапазоны",
			fields: "1,3-5,7,9-10",
			want: []FieldRange{
				{Start: 1, End: 1},
				{Start: 3, End: 5},
				{Start: 7, End: 7},
				{Start: 9, End: 10},
			},
			wantErr: false,
		},
		{
			name:   "с пробелами",
			fields: " 1 , 3-5 , 7 ",
			want: []FieldRange{
				{Start: 1, End: 1},
				{Start: 3, End: 5},
				{Start: 7, End: 7},
			},
			wantErr: false,
		},
		{
			name:    "пустая строка",
			fields:  "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "неверный диапазон",
			fields:  "1-3-5",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "отрицательное число",
			fields:  "-1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "начало больше конца",
			fields:  "5-3",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFieldRanges(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFieldRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !fieldRangesEqual(got, tt.want) {
				t.Errorf("parseFieldRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFields(t *testing.T) {
	tests := []struct {
		name   string
		ranges []FieldRange
		want   []int
	}{
		{
			name: "отдельные поля",
			ranges: []FieldRange{
				{Start: 1, End: 1},
				{Start: 3, End: 3},
				{Start: 5, End: 5},
			},
			want: []int{1, 3, 5},
		},
		{
			name: "диапазоны",
			ranges: []FieldRange{
				{Start: 1, End: 3},
				{Start: 5, End: 7},
			},
			want: []int{1, 2, 3, 5, 6, 7},
		},
		{
			name: "пересекающиеся диапазоны",
			ranges: []FieldRange{
				{Start: 1, End: 5},
				{Start: 3, End: 7},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "дублирующиеся поля",
			ranges: []FieldRange{
				{Start: 1, End: 1},
				{Start: 1, End: 1},
				{Start: 3, End: 5},
			},
			want: []int{1, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFields(tt.ranges)
			if !intSlicesEqual(got, tt.want) {
				t.Errorf("getFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		config Config
		want   string
		output bool
	}{
		{
			name: "выбор полей 1 и 3",
			line: "a\tb\tc\td",
			config: Config{
				Delimiter:  "\t",
				FieldRanges: []FieldRange{
					{Start: 1, End: 1},
					{Start: 3, End: 3},
				},
			},
			want:   "a\tc",
			output: true,
		},
		{
			name: "диапазон полей 2-4",
			line: "a\tb\tc\td\te",
			config: Config{
				Delimiter:  "\t",
				FieldRanges: []FieldRange{
					{Start: 2, End: 4},
				},
			},
			want:   "b\tc\td",
			output: true,
		},
		{
			name: "поле за границами",
			line: "a\tb",
			config: Config{
				Delimiter:  "\t",
				FieldRanges: []FieldRange{
					{Start: 1, End: 1},
					{Start: 5, End: 5},
				},
			},
			want:   "a",
			output: true,
		},
		{
			name: "без полей - выводим всю строку",
			line: "a\tb\tc",
			config: Config{
				Delimiter:   "\t",
				FieldRanges: []FieldRange{},
			},
			want:   "a\tb\tc",
			output: true,
		},
		{
			name: "флаг -s, строка без разделителя",
			line: "abc",
			config: Config{
				Delimiter:  "\t",
				Separated:  true,
				FieldRanges: []FieldRange{
					{Start: 1, End: 1},
				},
			},
			want:   "",
			output: false,
		},
		{
			name: "флаг -s, строка с разделителем",
			line: "a\tb",
			config: Config{
				Delimiter:  "\t",
				Separated:  true,
				FieldRanges: []FieldRange{
					{Start: 1, End: 1},
				},
			},
			want:   "a",
			output: true,
		},
		{
			name: "другой разделитель",
			line: "a,b,c,d",
			config: Config{
				Delimiter:  ",",
				FieldRanges: []FieldRange{
					{Start: 1, End: 1},
					{Start: 4, End: 4},
				},
			},
			want:   "a,d",
			output: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, output := processLine(tt.line, tt.config)
			if got != tt.want || output != tt.output {
				t.Errorf("processLine() = (%v, %v), want (%v, %v)", got, output, tt.want, tt.output)
			}
		})
	}
}

// Вспомогательные функции для сравнения

func fieldRangesEqual(a, b []FieldRange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Start != b[i].Start || a[i].End != b[i].End {
			return false
		}
	}
	return true
}

func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

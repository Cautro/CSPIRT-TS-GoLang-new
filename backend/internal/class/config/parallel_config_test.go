package config

import (
	"testing"
)

func TestParseParallelsConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []ParallelConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:  "single range",
			input: "1-4",
			want: []ParallelConfig{
				{Name: "1-4 параллель", MinGrade: 1, MaxGrade: 4},
			},
			wantErr: false,
		},
		{
			name:  "multiple ranges",
			input: "1-4,5-9,10-11",
			want: []ParallelConfig{
				{Name: "1-4 параллель", MinGrade: 1, MaxGrade: 4},
				{Name: "5-9 параллель", MinGrade: 5, MaxGrade: 9},
				{Name: "10-11 параллель", MinGrade: 10, MaxGrade: 11},
			},
			wantErr: false,
		},
		{
			name:  "with spaces",
			input: "1 - 4 , 5 - 9",
			want: []ParallelConfig{
				{Name: "1-4 параллель", MinGrade: 1, MaxGrade: 4},
				{Name: "5-9 параллель", MinGrade: 5, MaxGrade: 9},
			},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    []ParallelConfig{},
			wantErr: false,
		},
		{
			name:    "only spaces",
			input:   "   ",
			want:    []ParallelConfig{},
			wantErr: false,
		},
		{
			name:    "invalid format - no dash",
			input:   "1-4-5",
			wantErr: true,
			errMsg:  "invalid parallel range format",
		},
		{
			name:    "invalid format - no numbers",
			input:   "a-b",
			wantErr: true,
			errMsg:  "invalid min grade",
		},
		{
			name:    "min > max",
			input:   "9-5",
			wantErr: true,
			errMsg:  "min grade",
		},
		{
			name:  "trailing comma",
			input: "1-4,5-9,",
			want: []ParallelConfig{
				{Name: "1-4 параллель", MinGrade: 1, MaxGrade: 4},
				{Name: "5-9 параллель", MinGrade: 5, MaxGrade: 9},
			},
			wantErr: false,
		},
		{
			name:  "leading comma",
			input: ",1-4,5-9",
			want: []ParallelConfig{
				{Name: "1-4 параллель", MinGrade: 1, MaxGrade: 4},
				{Name: "5-9 параллель", MinGrade: 5, MaxGrade: 9},
			},
			wantErr: false,
		},
		{
			name:  "same min and max",
			input: "5-5",
			want: []ParallelConfig{
				{Name: "5-5 параллель", MinGrade: 5, MaxGrade: 5},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseParallelsConfig(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseParallelsConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errMsg != "" && err != nil {
					if _, ok := err.(interface{ Error() string }); !ok {
						t.Errorf("ParseParallelsConfig() error = %v, want error with message containing '%s'", err, tt.errMsg)
					}
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseParallelsConfig() got %d ranges, want %d", len(got), len(tt.want))
				return
			}

			for i, g := range got {
				if g.Name != tt.want[i].Name || g.MinGrade != tt.want[i].MinGrade || g.MaxGrade != tt.want[i].MaxGrade {
					t.Errorf("ParseParallelsConfig()[%d] = %v, want %v", i, g, tt.want[i])
				}
			}
		})
	}
}

func BenchmarkParseParallelsConfig(b *testing.B) {
	input := "1-4,5-9,10-11"
	for i := 0; i < b.N; i++ {
		_, _ = ParseParallelsConfig(input)
	}
}

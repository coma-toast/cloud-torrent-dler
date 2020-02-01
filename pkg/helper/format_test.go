package helper

import "testing"

func Test_sanitizeText(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "spaces",
			args: args{input: "Rick And Morty"},
			want: "rick and morty",
		},
		{
			name: "foldername",
			args: args{input: "Rick_And_Morty S01E01 720p-BluRay 5.1 BONE"},
			want: "rick and morty s01e01 720p bluray 5 1 bone",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeText(tt.args.input); got != tt.want {
				t.Errorf("sanitizeText() = %v, want %v", got, tt.want)
			}
		})
	}
}

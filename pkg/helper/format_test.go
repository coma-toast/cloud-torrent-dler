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
		{
			name: "webrip",
			args: args{input: "The 100 S07E13 1080p CW WEBRip AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 h264 btn",
		},
		{
			name: "web dl",
			args: args{input: "The 100 S07E13 1080p CW WEB-DL AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 h264 btn",
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

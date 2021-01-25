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
			want: "rick and morty s01e01 720p-bluray 5.1 bone",
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
		{
			name: "H264",
			args: args{input: "The 100 S07E13 1080p CW WEB-DL AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 h264 btn",
		},
		{
			name: "H.264",
			args: args{input: "The.100.S07E13.1080p.CW.WEB-DL.AAC2.0.H.264-BTN"},
			want: "the.100.s07e13.1080p.cw.aac2.0.h.264-btn",
		},
		{
			name: "[]",
			args: args{input: "23.Walks.2020.1080p.WEB-DL.DD5.1.H.264-EVO[TGx]"},
			want: "23.walks.2020.1080p.dd5.1.h.264-evotgx",
		},
		{
			name: "lots of dots",
			args: args{input: "Paw.Patrol.S07E30.Moto.Pups.Pups.vs.the.Ruff-Ruff.Pack.1080p.NICK.WEB-DL.AAC2.0.H.264-LAZY"},
			want: "paw.patrol.s07e30.moto.pups.pups.vs.the.ruff-ruff.pack.1080p.nick.aac2.0.h.264-lazy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeText(tt.args.input); got != tt.want {
				t.Errorf("\n\nsanitizeText()\n  got:\n    %v\n  want:\n    %v\n\n", got, tt.want)
			}
		})
	}
}

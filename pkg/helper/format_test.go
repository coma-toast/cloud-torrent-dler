package helper

import (
	"testing"
)

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
			name: "webrip",
			args: args{input: "The 100 S07E13 1080p CW WEBRip AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 btn",
		},
		{
			name: "web dl",
			args: args{input: "The 100 S07E13 1080p CW WEB-DL AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 btn",
		},
		{
			name: "H264",
			args: args{input: "The 100 S07E13 1080p CW WEB-DL AAC2 0 H264 BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 btn",
		},
		{
			name: "H.264",
			args: args{input: "The.100.S07E13.1080p.CW.WEB-DL.AAC2.0.H.264-BTN"},
			want: "the 100 s07e13 1080p cw aac2 0 btn",
		},
		{
			name: "[]",
			args: args{input: "23.Walks.2020.1080p.WEB-DL.DD5.1.H.264-EVO[TGx]"},
			want: "23 walks 2020 1080p dd5 1 evotgx",
		},
		{
			name: "lots of dots",
			args: args{input: "Paw.Patrol.S07E30.Moto.Pups.Pups.vs.the.Ruff-Ruff.Pack.1080p.NICK.WEB-DL.AAC2.0.H.264-LAZY"},
			want: "paw patrol s07e30 moto pups pups vs the ruff ruff pack 1080p nick aac2 0 lazy",
		},
		{
			name: "SanitizeText for ShowRSS",
			args: args{input: "The Curse of Oak Island S08E11 1080p WEB H264 WHOSNEXT"},
			want: "the curse of oak island s08e11 1080p web whosnext",
		},
		{
			name: "SanitizeText for Seedr",
			args: args{input: "the.curse.of.oak.island.s08e11.1080p.web.h264-whosnext"},
			want: "the curse of oak island s08e11 1080p web whosnext",
		},
		{
			name: "This is the folder name in the torrent file",
			args: args{input: "The.Curse.of.Oak.Island.S08E11.1080p.WEB.H264-WHOSNEXT[rarbg]"},
			want: "the curse of oak island s08e11 1080p web whosnext",
		},
		{
			name: "ShowRSS Name",
			args: args{input: "Paw Patrol S07E30 Moto Pups Pups vs the Ruff Ruff Pack 1080p NICK WEBRip AAC2 0 H264 LAZY"},
			want: "paw patrol s07e30 moto pups pups vs the ruff ruff pack 1080p nick aac2 0 lazy",
		},
		{
			name: "Workin Moms",
			args: args{input: "Workin.Moms.S05E03.1080p.WEBRip.x264-BAE[rarbg]"},
			want: "workin moms s05e03 1080p bae",
		},
		{
			name: "Workin Moms another way",
			args: args{input: "Workin Moms S05E04 1080p WEBRip x264 BAE"},
			want: "workin moms s05e04 1080p bae",
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

func TestSanitizePath(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Seedr Root Path",
			args: args{input: "The.Curse.of.Oak.Island.S08E11.1080p.WEB.H264-WHOSNEXT[rarbg]"},
			want: "the.curse.of.oak.island.s08e11.1080p.web.h264-whosnextrarbg",
		},
		{
			name: "Seedr Subfolder Path",
			args: args{input: "Shows/NotKids/The.Curse.of.Oak.Island.S08E11.1080p.WEB.H264-WHOSNEXT[rarbg]"},
			want: "Shows/NotKids/the.curse.of.oak.island.s08e11.1080p.web.h264-whosnextrarbg",
		},
		{
			name: "Seedr Subfolder Path2",
			args: args{input: "Shows/Kids/Paw.Patrol.S07E30.Moto.Pups.Pups.vs.the.Ruff-Ruff.Pack.1080p.NICK.WEBRip.AAC2.0.H264-LAZY[rarbg]"},
			want: "Shows/NotKids/the.curse.of.oak.island.s08e11.1080p.web.h264-whosnextrarbg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizePath(tt.args.input); got != tt.want {
				t.Errorf("\n\nsanitizePath()\n  got:\n    %v\n  want:\n    %v\n\n", got, tt.want)
			}
		})
	}
}

func TestSanitizeBoth(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Seedr Name",
			args: args{input: "Paw.Patrol.S07E30.Moto.Pups.Pups.vs.the.Ruff-Ruff.Pack.1080p.NICK.WEB-DL.AAC2.0.H.264-LAZY"},
			want: "paw patrol s07e30 moto pups pups vs the ruff ruff pack 1080p nick aac2 0 lazy",
		},
		{
			name: "ShowRSS Name",
			args: args{input: "Paw Patrol S07E30 Moto Pups Pups vs the Ruff Ruff Pack 1080p NICK WEBRip AAC2 0 H264 LAZY"},
			want: "paw patrol s07e30 moto pups pups vs the ruff ruff pack 1080p nick aac2 0 lazy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizePath(tt.args.input); got != tt.want {
				t.Errorf("\n\nsanitizePath()\n  got:\n    %v\n  want:\n    %v\n\n", got, tt.want)
			}
			if got := SanitizeText(tt.args.input); got != tt.want {
				t.Errorf("\n\nsanitizeText()\n  got:\n    %v\n  want:\n    %v\n\n", got, tt.want)
			}
		})
	}
}

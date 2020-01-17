package main

import (
	"reflect"
	"testing"
)

func Test_findAllToDownload(t *testing.T) {
	seedrInstance := mockSeedrInstance{}
	type args struct {
		instance SeedrInstance
		path     string
		ftp      bool
	}
	tests := []struct {
		name    string
		args    args
		want    []DownloadItem
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				instance: seedrInstance,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findAllToDownload(tt.args.instance, tt.args.path, tt.args.ftp)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAllToDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAllToDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/infobloxopen/atlas-openapiv2-patch/internal/descriptor"
)

const (
	succeed                    = "\u2713"
	failed                     = "\u2717"
	red                        = "\033[31m"
	green                      = "\033[32m"
	yellow                     = "\033[33m"
	reset                      = "\033[0m"
	atlasPatchInputSwaggerFile = "../../internal/testdata/atlaspatch.in.swagger.json"
	jeepSvcInputSwaggerFile    = "../../internal/testdata/jeepsvc.in.swagger.json"
)

func deepCompare(t *testing.T, file1, file2 string) (bool, error) {
	const chunkSize = 64000

	f1, err := os.Open(file1)
	if err != nil {
		return false, fmt.Errorf("cannot open %q file, error: %v", file1, err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, fmt.Errorf("cannot open %q file, error: %v", file2, err)
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true, nil
			} else if err1 == io.EOF || err2 == io.EOF {
				return false, nil
			} else {
				return false, fmt.Errorf("unexpected (not EOF) error "+
					"when reading files (1, 2): %v, %v", err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			t.Logf("Name of file1 %s, Generated contents:\n%s\n-------------\n\nName of file2 name (Wanted contents) %s\n",
				file1, string(b1), file2)
			return false, nil
		}
	}
}

func createFiles(inFile string, fileNames []string) error {
	var f []byte
	f, err := ioutil.ReadFile(inFile)
	if err != nil {
		return err
	}

	for _, file := range fileNames {
		err = ioutil.WriteFile(file, f, os.FileMode(0666))
		return err
	}
	return nil
}

func deleteFiles(fileNames []string) error {
	for _, file := range fileNames {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func Test_run(t *testing.T) {
	type args struct {
		withPrivate           bool
		withCustomAnnotations bool
		withPostResponse      int
		withPutResponse       int
		withPatchResponse     int
		withDeleteResponse    int
		files                 []string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		inputFile      string
		wantFile       string
		generatedFiles []string
	}{
		{
			name:    "without custom flags",
			wantErr: false,
			args: args{
				withPrivate:           false,
				withCustomAnnotations: false,
				files:                 []string{"../../internal/testdata/atlaspatch.emitted.swagger.json"},
			},
			inputFile:      atlasPatchInputSwaggerFile,
			wantFile:       "../../internal/testdata/atlaspatch.wanted.swagger.json",
			generatedFiles: []string{"../../internal/testdata/atlaspatch.emitted.swagger.json"},
		},
		{
			name:    "with custom annotation flag",
			wantErr: false,
			args: args{
				withPrivate:           false,
				withCustomAnnotations: true,
				files:                 []string{"../../internal/testdata/atlaspatch.emitted.swagger.json"},
			},
			inputFile:      atlasPatchInputSwaggerFile,
			wantFile:       "../../internal/testdata/atlaspatch.wanted.swagger.json",
			generatedFiles: []string{"../../internal/testdata/atlaspatch.emitted.swagger.json"},
		},
		{
			name:    "with private flags",
			wantErr: false,
			args: args{
				withPrivate:           true,
				withCustomAnnotations: false,
				files:                 []string{"../../internal/testdata/atlaspatch.emitted.swagger.json"},
			},
			inputFile:      atlasPatchInputSwaggerFile,
			wantFile:       "../../internal/testdata/atlaspatch.wanted.private.swagger.json",
			generatedFiles: []string{"../../internal/testdata/atlaspatch.emitted.private.swagger.json"},
		},
		{
			name:    "with custom HTTP response codes 1 - use non-default response codes",
			wantErr: false,
			args: args{
				withPostResponse:   302,
				withPutResponse:    303,
				withPatchResponse:  304,
				withDeleteResponse: 305,
				files:              []string{"../../internal/testdata/atlaspatch.emitted.customresponses1.swagger.json"},
			},
			inputFile:      atlasPatchInputSwaggerFile,
			wantFile:       "../../internal/testdata/atlaspatch.wanted.customresponses1.swagger.json",
			generatedFiles: []string{"../../internal/testdata/atlaspatch.emitted.customresponses1.swagger.json"},
		},
		{
			name:    "with custom HTTP response codes 2 - use 200 response codes",
			wantErr: false,
			args: args{
				withPostResponse:   200,
				withPutResponse:    200,
				withPatchResponse:  200,
				withDeleteResponse: 200,
				files:              []string{"../../internal/testdata/atlaspatch.emitted.customresponses2.swagger.json"},
			},
			inputFile:      atlasPatchInputSwaggerFile,
			wantFile:       "../../internal/testdata/atlaspatch.wanted.customresponses2.swagger.json",
			generatedFiles: []string{"../../internal/testdata/atlaspatch.emitted.customresponses2.swagger.json"},
		},
		{
			name:    "jeep svc file with custom HTTP response codes - use 200 response codes",
			wantErr: false,
			args: args{
				withPostResponse:   200,
				withPutResponse:    200,
				withPatchResponse:  200,
				withDeleteResponse: 200,
				files:              []string{"../../internal/testdata/jeepsvc.emitted.customresponses.swagger.json"},
			},
			inputFile:      jeepSvcInputSwaggerFile,
			wantFile:       "../../internal/testdata/jeepsvc.wanted.customresponses.swagger.json",
			generatedFiles: []string{"../../internal/testdata/jeepsvc.emitted.customresponses.swagger.json"},
		},
	}

	reg := descriptor.NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg.SetPrivateOperations(tt.args.withPrivate)
			reg.SetCustomAnnotations(tt.args.withCustomAnnotations)
			if tt.args.withPostResponse > 0 {
				reg.SetPostResponse(tt.args.withPostResponse)
			}
			if tt.args.withPutResponse > 0 {
				reg.SetPutResponse(tt.args.withPutResponse)
			}
			if tt.args.withPatchResponse > 0 {
				reg.SetPatchResponse(tt.args.withPatchResponse)
			}
			if tt.args.withDeleteResponse > 0 {
				reg.SetDeleteResponse(tt.args.withDeleteResponse)
			}
			err := createFiles(tt.inputFile, tt.args.files)
			defer deleteFiles(tt.generatedFiles)
			defer deleteFiles(tt.args.files)
			if err != nil {
				t.Fatal(err)
			}
			err = run(reg, tt.args.files)
			if err != nil {
				t.Errorf("unable to apply atlas patch, %v", err)
				return
			}

			isEqual, err := deepCompare(t, tt.generatedFiles[0], tt.wantFile)
			if err != nil {
				t.Errorf("Emitted vs wanted files content comparison error: %v", err)
				return
			}

			if !isEqual {
				t.Errorf("\t%s Emitted swagger JSON file is not equal to wanted one, compare with: "+
					"\n"+yellow+" \n\ndiff %s %s\n\n "+reset, failed, tt.args.files, tt.wantFile)
				return
			}
			t.Logf("\t%s %s test is passed", succeed, tt.name)
		})
	}
}

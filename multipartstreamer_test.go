package multipartstreamer

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func TestMultipart(t *testing.T) {
	path, _ := os.Getwd()
	file := filepath.Join(path, "multipartstreamer.go")
	stat, _ := os.Stat(file)

	ms := New()
	err := ms.Write("file", file, map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("Error writing fields: %s", err)
	}

	diff := ms.Len() - stat.Size()
	if diff != 398 {
		t.Error("Unexpected multipart size")
	}

	data, err := ioutil.ReadAll(ms.GetReader())
	if err != nil {
		t.Fatalf("Error reading multipart data: %s", err)
	}

	buf := bytes.NewBuffer(data)
	reader := multipart.NewReader(buf, ms.Boundary())

	part, err := reader.NextPart()
	if err != nil {
		t.Fatalf("Expected form field: %s", err)
	}

	if str := part.FileName(); str != "" {
		t.Errorf("Unexpected filename: %s", str)
	}

	if str := part.FormName(); str != "a" {
		t.Errorf("Unexpected form name: %s", str)
	}

	if by, _ := ioutil.ReadAll(part); string(by) != "b" {
		t.Errorf("Unexpected form value: %s", string(by))
	}

	part, err = reader.NextPart()
	if err != nil {
		t.Fatalf("Expected file field: %s", err)
	}

	if str := part.FileName(); str != file {
		t.Errorf("Unexpected filename: %s", str)
	}

	if str := part.FormName(); str != "file" {
		t.Errorf("Unexpected form name: %s", str)
	}

	src, _ := ioutil.ReadFile(file)
	if by, _ := ioutil.ReadAll(part); string(by) != string(src) {
		t.Errorf("Unexpected file value")
	}

	part, err = reader.NextPart()
	if err != io.EOF {
		t.Errorf("Unexpected 3rd part: %s", part)
	}
}

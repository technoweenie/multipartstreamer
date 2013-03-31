package multipartstreamer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type MultipartStreamer struct {
	ContentType   string
	bodyBuffer    *bytes.Buffer
	bodyWriter    *multipart.Writer
	closeBuffer   *bytes.Buffer
	reader        io.Reader
	contentLength int64
}

// Helps you build multipart for large files without reading the file until the
// multipart reader is being read.  It does this by creating the file field last
// and using an io.MultiReader to combine the multipart.Reader with the file
// handle.  The trailing boundary is manually added in another buffer.
func New() (m *MultipartStreamer) {
	m = &MultipartStreamer{bodyBuffer: bytes.NewBufferString("")}

	m.bodyWriter = multipart.NewWriter(m.bodyBuffer)
	boundary := m.bodyWriter.Boundary()
	m.ContentType = "multipart/form-data; boundary=" + boundary

	closeBoundary := fmt.Sprintf("\r\n--%s--\r\n", boundary)
	m.closeBuffer = bytes.NewBufferString(closeBoundary)

	return
}

// Writes form fields to the multipart.Writer.
//
// fields   - A map of form field keys and values.
func (m *MultipartStreamer) WriteFields(fields map[string]string) error {
	var err error

	for key, value := range fields {
		err = m.bodyWriter.WriteField(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Prepares a file to be written to the multipart.Writer.
//
// key - The name of the field for the file data.
//
// filename - The name of the file to upload.
func (m *MultipartStreamer) WriteFile(key, filename string) error {
	fh, err := os.Open(filename)
	if err != nil {
		return err
	}

	stat, err := fh.Stat()
	if err != nil {
		return err
	}

	return m.WriteReader(key, filepath.Base(filename), stat.Size(), fh)
}

func (m *MultipartStreamer) WriteReader(key, filename string, size int64, reader io.Reader) (err error) {
	m.reader = reader
	m.contentLength = size

	_, err = m.bodyWriter.CreateFormFile(key, filepath.Base(filename))
	return
}

// Sets up the http.Request body, and some crucial HTTP headers.
func (m *MultipartStreamer) SetupRequest(req *http.Request) {
	req.Body = m.GetReader()
	req.Header.Add("Content-Type", m.ContentType)
	req.ContentLength = m.Len()
}

func (m *MultipartStreamer) Boundary() string {
	return m.bodyWriter.Boundary()
}

// Calculates the byte size of the multipart content.
func (m *MultipartStreamer) Len() int64 {
	return m.contentLength + int64(m.bodyBuffer.Len()) + int64(m.closeBuffer.Len())
}

// Gets an io.ReadCloser for passing to an http.Request.
func (m *MultipartStreamer) GetReader() io.ReadCloser {
	reader := io.MultiReader(m.bodyBuffer, m.reader, m.closeBuffer)
	return ioutil.NopCloser(reader)
}

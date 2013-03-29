package multipartstreamer

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "mime/multipart"
  "net/http"
  "os"
)

type MultipartStreamer struct {
  ContentType   string
  bodyBuffer    *bytes.Buffer
  bodyWriter    *multipart.Writer
  closeBuffer   *bytes.Buffer
  fileHandle    *os.File
  contentLength int64
}

func New() (m *MultipartStreamer) {
  m = &MultipartStreamer{bodyBuffer: bytes.NewBufferString("")}

  m.bodyWriter = multipart.NewWriter(m.bodyBuffer)
  boundary := m.bodyWriter.Boundary()
  m.ContentType = "multipart/form-data; boundary=" + boundary

  closeBoundary := fmt.Sprintf("\r\n--%s--\r\n", boundary)
  m.closeBuffer = bytes.NewBufferString(closeBoundary)

  return
}

func (m *MultipartStreamer) Write(key, filename string, fields map[string]string) (err error) {
  err = m.WriteFields(fields)
  if err != nil {
    return
  }

  err = m.WriteFile(key, filename)

  return
}

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

func (m *MultipartStreamer) WriteFile(key, filename string) (err error) {
  _, err = m.bodyWriter.CreateFormFile(key, filename)
  if err != nil {
    return
  }

  m.fileHandle, err = os.Open(filename)
  if err != nil {
    return
  }

  var stat os.FileInfo
  stat, err = m.fileHandle.Stat()
  m.contentLength = stat.Size()

  return
}

func (m *MultipartStreamer) SetupRequest(req *http.Request) {
  req.Body = m.GetReader()
  req.Header.Add("Content-Type", m.ContentType)
  req.ContentLength = m.Len()
}

func (m *MultipartStreamer) Len() int64 {
  return m.contentLength + int64(m.bodyBuffer.Len()) + int64(m.closeBuffer.Len())
}

func (m *MultipartStreamer) GetReader() io.ReadCloser {
  reader := io.MultiReader(m.bodyBuffer, m.fileHandle, m.closeBuffer)
  return ioutil.NopCloser(reader)
}

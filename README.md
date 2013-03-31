# multipartstreamer

Helps you build multipart for large files without reading the file until the
multipart reader is being read.  It does this by creating the file field last
and using an io.MultiReader to combine the multipart.Reader with the file
handle.  The trailing boundary is manually added in another buffer.

The reason you don't want to just use the built-in multipart.Reader is that it
will read the whole file into the buffer to build the multipart content.

```go
package main

import (
  "github.com/technoweenie/multipartstreamer.go"
  "net/http"
)

func main() {
  ms := multipartstreamer.New()

  ms.WriteFields(map[string]string{
    "key":			"some-key",
    "AWSAccessKeyId":	"ABCDEF",
    "acl":			"some-acl",
  })

  // Add any io.Reader to the multipart.Reader.
  ms.WriteReader("file", "filename", some_ioReader, size)

  // Shortcut for adding local file.
  ms.WriteFile("file", "path/to/file")

  req, _ := http.NewRequest("POST", "someurl", nil)
  ms.SetupRequest(req)

  res, _ := http.DefaultClient.Do(req)
}
```

One limitation: You can only write a single file.

## TODO

* Multiple files?

## Credits

Heavily inspired by James

https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/Zjg5l4nKcQ0
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
)

func main() {
	port := flag.String("p", "9100", "port to serve on")
	directory := flag.String("d", "media/", "the directory of static file to host")
	flag.Parse()

	fs := withLog(http.FileServer(http.Dir(*directory)).ServeHTTP)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			rangeVideo(w, r)
			return
		}
		if r.URL.Path == "/norange" {
			norange(w, r)
			return
		}
		fs(w, r)
	})

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

const (
	vname         = "./media/tomato-egg_stir-fry.mp4"
	sizePerRequst = 5 * 1000 * 1000 // 5MB/req
)

func openfile(name string) (*os.File, int64, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, 0, err
	}
	finfo, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, finfo.Size(), nil
}

func norange(w http.ResponseWriter, req *http.Request) {
	f, size, err := openfile(vname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "video/mp4")

	if req.Method != "HEAD" {
		io.CopyN(w, f, size)
	}
}

func rangeVideo(w http.ResponseWriter, req *http.Request) {
	f, size, err := openfile(vname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "video/mp4")

	rangeHeader := req.Header.Get("Range")
	// we can simply hint Chrome to send serial range requests for media file by
	//
	// if rangeHeader == "" {
	// 	w.Header().Set("Accept-Ranges", "bytes")
	// 	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	// 	w.WriteHeader(200)
	// 	fmt.Printf("hint browser to send range requests, total size: %d\n", size)
	// 	return
	// }
	//
	// but this not worked for Safari and Firefox
	if rangeHeader == "" {
		ra := httpRange{
			start:  0,
			length: sizePerRequst,
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", strconv.FormatInt(ra.length, 10))
		w.Header().Set("Content-Range", ra.contentRange(size))
		w.WriteHeader(http.StatusPartialContent)
		fmt.Printf("hint browser to send serial range requests, response 206, 0-%d/%d\n", sizePerRequst-1, size)
		if req.Method != "HEAD" {
			written, err := io.CopyN(w, f, ra.length)
			if written != ra.length {
				fmt.Printf("desired range size: %d, actual written: %d, err: %v\n\n", ra.length, written, err)
			}
		}
		return
	}

	// browser sends range request
	reqer := req.RemoteAddr
	fmt.Printf("\n%s request range %s\n", reqer, rangeHeader)
	ranges, err := parseRange(rangeHeader, size)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// multi-part requests are not supported
	if len(ranges) > 1 {
		http.Error(w, "unsuported multi-part", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	ra := ranges[0]
	if _, err := f.Seek(ra.start, io.SeekStart); err != nil {
		http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
		return
	}
	fmt.Printf("response range bytes %d-%d, %d KB\n", ra.start, ra.start+ra.length-1, ra.length/1024)
	sendSize := ra.length
	w.Header().Set("Content-Range", ra.contentRange(size))
	w.Header().Set("Accept-Ranges", "bytes")
	if w.Header().Get("Content-Encoding") == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}
	w.WriteHeader(http.StatusPartialContent)

	if req.Method != "HEAD" {
		written, err := io.CopyN(w, f, sendSize)
		if written != sendSize || err != nil {
			fmt.Printf("desired range size: %d, actual written: %d, err: %v\n\n", sendSize, written, err)
		} else {
			fmt.Println()
		}
	}
}

// --- httpRange and its funcs are ported from net/http fs.go

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
	//return fmt.Sprintf("bytes %d-%d/*", r.start, r.start+r.length-1)
}

var errNoOverlap = errors.New("invalid range: failed to overlap")

// parseRange parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				r.length = sizePerRequst
				if r.length > size-r.start {
					r.length = size - r.start
				}
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errNoOverlap
	}
	return ranges, nil
}

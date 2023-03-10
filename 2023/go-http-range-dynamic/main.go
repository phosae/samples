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

var isFakeDynamic bool

func main() {
	port := flag.String("p", "9100", "port to serve on")
	directory := flag.String("d", "media/", "the directory of static file to host")
	flag.BoolVar(&isFakeDynamic, "dynamic", false, "whether return a large enough Content-Length to browser")
	flag.Parse()

	fs := withLog(http.FileServer(http.Dir(*directory)).ServeHTTP)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			rangeVideo(w, r)
			return
		}
		fs(w, r)
	})

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

const (
	vpart1         = "./media/dun-dun-dance-part1.mp4"
	vfull          = "./media/dun-dun-dance.mp4"
	sizePerRequst  = 5 * 1000 * 1000    // 5MB/req
	largeEnoughLen = 1000 * 1000 * 1000 // 1GB
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

func rangeVideo(w http.ResponseWriter, req *http.Request) {
	f, size, err := openfile(vpart1)
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
		if isFakeDynamic {
			w.Header().Set("Content-Range", ra.contentRange(largeEnoughLen))
		} else {
			w.Header().Set("Content-Range", ra.contentRange(size))
		}

		w.WriteHeader(http.StatusPartialContent)
		fmt.Printf("hint browser to send serial range requests, response 206, 0-%d/%s\n", sizePerRequst-1, w.Header().Get("Content-Range"))
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
	// if isFakeDynamic {
	// 	ranges, err = parseRange(rangeHeader, largeEnoughLen)
	// }
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

	if ra.start+sizePerRequst > size && ra.length > 1024*1024 /* try trick the tail verify */ {
		fmt.Printf("part1 size %d, range start %d size %d,open full file\n", size, ra.start, ra.length)
		f, size, err = openfile(vfull)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer f.Close()
	}

	if _, err := f.Seek(ra.start, io.SeekStart); err != nil {
		if ra.start+ra.length == largeEnoughLen { // trick the tail verify
			_, err = f.Seek(size-ra.length, io.SeekStart)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}
	fmt.Printf("response range bytes %d-%d, %d KB\n", ra.start, ra.start+ra.length-1, ra.length/1024)
	sendSize := ra.length

	if isFakeDynamic {
		w.Header().Set("Content-Range", ra.contentRange(largeEnoughLen))
	} else {
		w.Header().Set("Content-Range", ra.contentRange(size))
	}

	w.Header().Set("Accept-Ranges", "bytes")
	if w.Header().Get("Content-Encoding") == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}
	w.WriteHeader(http.StatusPartialContent)

	if req.Method != "HEAD" {
		written, err := io.CopyN(w, f, sendSize)
		if written != sendSize || err != nil {
			fmt.Println(ra, size)
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
				//continue, since server may attempt to return a largeEnoughLen, errNoOverlap never happen
			}
			r.start = i
			if end == "" {
				r.length = sizePerRequst
				if r.length > size-r.start && !noOverlap {
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

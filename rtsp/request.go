package rtsp

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
)

const (
	rtspProtocol10           = "RTSP/1.0"
	requestMaxMethodLength   = 64
	requestMaxURLLength      = 2048
	requestMaxProtocolLength = 64
)

// Method is the method of a RTSP request.
type Method string

// standard methods
const (
	Announce     Method = "ANNOUNCE"
	Describe     Method = "DESCRIBE"
	GetParameter Method = "GET_PARAMETER"
	Options      Method = "OPTIONS"
	Pause        Method = "PAUSE"
	Play         Method = "PLAY"
	Record       Method = "RECORD"
	Setup        Method = "SETUP"
	SetParameter Method = "SET_PARAMETER"
	Teardown     Method = "TEARDOWN"
	Flush        Method = "FLUSH"
	Post         Method = "POST"
	Get          Method = "GET"
)

// Request is a RTSP request.
type Request struct {
	// request method
	Method Method

	// request url
	URL *URL

	Path string

	Query string

	// map of header values
	Header Header

	// optional body
	Body []byte
}

// Read reads a request.
func (req *Request) Read(rb *bufio.Reader) error {
	byts, err := readBytesLimited(rb, ' ', requestMaxMethodLength)
	if err != nil {
		return err
	}
	req.Method = Method(byts[:len(byts)-1])

	if req.Method == "" {
		return fmt.Errorf("empty method")
	}

	byts, err = readBytesLimited(rb, ' ', requestMaxURLLength)
	if err != nil {
		return err
	}
	rawURL := string(byts[:len(byts)-1])

	ur, err := ParseURL(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL (%v)", rawURL)
	}
	req.URL = ur
	pathAndQuery, ok := req.URL.RTSPPath()
	if !ok {
		return fmt.Errorf("URL path and query not valid")
	}
	req.Path, req.Query = PathSplitQuery(pathAndQuery)

	byts, err = readBytesLimited(rb, '\r', requestMaxProtocolLength)
	if err != nil {
		return err
	}
	proto := string(byts[:len(byts)-1])

	if proto != rtspProtocol10 {
		return fmt.Errorf("expected '%s', got '%s'", rtspProtocol10, proto)
	}

	err = readByteEqual(rb, '\n')
	if err != nil {
		return err
	}

	err = req.Header.read(rb)
	if err != nil {
		return err
	}

	err = (*body)(&req.Body).read(req.Header, rb)
	if err != nil {
		return err
	}

	return nil
}

// Write writes a request.
func (req Request) Write(bw *bufio.Writer) error {
	urStr := req.URL.CloneWithoutCredentials().String()
	_, err := bw.Write([]byte(string(req.Method) + " " + urStr + " " + rtspProtocol10 + "\r\n"))
	if err != nil {
		return err
	}

	if len(req.Body) != 0 {
		req.Header["Content-Length"] = HeaderValue{strconv.FormatInt(int64(len(req.Body)), 10)}
	}

	err = req.Header.write(bw)
	if err != nil {
		return err
	}

	err = body(req.Body).write(bw)
	if err != nil {
		return err
	}

	return bw.Flush()
}

// String implements fmt.Stringer.
func (req Request) String() string {
	buf := bytes.NewBuffer(nil)
	req.Write(bufio.NewWriter(buf))
	return buf.String()
}

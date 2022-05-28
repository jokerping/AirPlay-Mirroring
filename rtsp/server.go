package rtsp

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"bufio"
	"net"
)

const (
	serverConnReadBufferSize  = 4096
	serverConnWriteBufferSize = 4096
)

type Handler interface {
	Handle(conn *Conn, req *Request) (*Response, error)
	OnRequest(conn *Conn, request *Request)
	OnResponse(conn *Conn, resp *Response)
	OnConnOpen(conn *Conn)
}

type Server struct {
	h                  Handler
	bw                 *bufio.Writer
	br                 *bufio.Reader
	Eiv                []byte
	Ekey               []byte // setup阶段传输的key，用于解密视频流
	StreamConnectionID int64  //视频流连接ID，用于解密视频流
	KeyMessage         []byte //用于解码aes-key
	EcdhShared         []byte //pairing阶段计算出的curve25519 共享密钥，用于解密视频
	VideoAESKey        []byte //解密视频的key
	VideoAESIv         []byte
	TimePort           uint64 //客户端NTP端口，用于对时
}

type Conn struct {
	c net.Conn
}

func (c *Conn) NetConn() net.Conn {
	return c.c
}

func (c *Conn) Close() error {
	return c.c.Close()
}

func (c *Conn) SetNetConn(conn net.Conn) {
	c.c = conn
}
func RunRtspServer(handlers Handler) (err error) {
	Session = &Server{
		h: handlers,
	}
	l, err := net.Listen("tcp4", config.Config.Port)
	if err == nil {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				global.Debug.Println("Error accepting: ", err.Error())
				return err
			}
			rConn := &Conn{
				c: conn,
			}
			Session.h.OnConnOpen(rConn)
			go Session.handleRstpConnection(rConn)
		}
	}
	return err
}

func (s Server) handleRstpConnection(conn *Conn) {
	defer conn.Close()

	s.br = bufio.NewReaderSize(conn.NetConn(), serverConnReadBufferSize)
	s.bw = bufio.NewWriterSize(conn.NetConn(), serverConnWriteBufferSize)

	for {
		request, err := parseRequest(s.br)
		if err != nil {
			global.Debug.Printf("Error parsing RSTP request %v \n", err)
			return
		}
		s.h.OnRequest(conn, request)
		response, err := s.h.Handle(conn, request)
		if err != nil {
			global.Debug.Printf("Error handling RSTP request %v \n", err)
			return
		}
		err = s.flushResponse(conn, request, response)
		if err != nil {
			global.Debug.Printf("Error flusing RSTP response %v \n", err)
			return
		}
	}
}

func (s *Server) flushResponse(conn *Conn, req *Request, resp *Response) error {
	if resp.Header == nil {
		resp.Header = make(Header)
	}
	resp.Header["CSeq"] = req.Header["CSeq"]
	resp.Header["Server"] = HeaderValue{"AirTunes/366.0"}
	s.h.OnResponse(conn, resp)
	return resp.Write(s.bw)
}

func parseRequest(br *bufio.Reader) (*Request, error) {
	var req Request
	var err error
	if err = req.Read(br); err != nil {
		return nil, err
	}
	return &req, nil
}

var Session *Server

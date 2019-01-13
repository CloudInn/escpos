package escpos

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
)

const (
	// DefaultEndpoint is the default server endpoint for ePOS printers.
	DefaultEndpoint = "/cgi-bin/epos/service.cgi"
)

// Server wrap
type Server struct {
	p      *Printer
	w      *bufio.Writer
	logger func(string, ...interface{})
}

// NewServer creates a new ePOS server.
func NewServer(w io.Writer, opts ...ServerOption) (*Server, error) {
	var err error

	// create printer
	p, err := NewPrinter(w)
	if err != nil {
		return nil, err
	}

	s := &Server{
		p: p,
		w: bufio.NewWriter(p),
	}

	// apply opts
	for _, o := range opts {
		err = o(s)
		if err != nil {
			return nil, err
		}
	}

	if s.logger == nil {
		s.logger = func(string, ...interface{}) {}
	}

	return s, nil
}

// ServeHTTP handles OPTIONS, Origin, and POST for an ePOS server.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.logger("%s %s", req.Method, req.URL)

	// send origin headers
	if origin := req.Header.Get("Origin"); origin != "" {
		res.Header().Set("Access-Control-Allow-Origin", origin)
		res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		res.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, If-Modified-Since, SOAPAction")
	}

	// stop if its options
	if req.Method == "OPTIONS" {
		return
	}

	// bail if not POST
	if req.Method != "POST" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// init printer
	s.p.Init()

	// end
	s.p.End()

	// flush writer
	s.w.Flush()

	// write soap response
	res.Header().Set("Content-Type", req.Header.Get("Content-Type"))
	fmt.Fprintf(res, soapBody, true, "")
}

const (
	// soapBody is a basic SOAP response body for an ePOS server response.
	soapBody = `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body xmlns:m="http://www.epson-pos.com/schemas/2011/03/epos-print">
	<m:response success="%t" code="%s" status="0"></m:response>
  </s:Body>
</s:Envelope>`
)

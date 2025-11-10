package http1

import (
	"bytes"
	"errors"
	"strconv"
	"unicode/utf8"
)

var ErrBodyTooSmall = errors.New("body too small")

var (
	CRLF  = []byte("\r\n")
	CRLF2 = []byte("\r\n\r\n")
	WS    = byte(' ')
)

type HttpRequestMessage struct {
	Method  []byte
	Target  []byte
	Version []byte
	Headers map[string][]byte
	Body    []byte
}

type ParseState struct {
	Msg       HttpRequestMessage
	Remaining []byte
}

func ParseRequestLine(s []byte) (ParseState, error) {
	requestLine, remaining, found := bytes.Cut(s, CRLF)

	if !found {
		return ParseState{}, errors.New("malformed http message")
	}

	split := bytes.SplitN(requestLine, []byte{WS}, 3)

	if len(split) != 3 {
		return ParseState{}, errors.New("malformed http request line")
	}

	method := split[0]
	target := split[1]
	version := split[2]

	return ParseState{
		HttpRequestMessage{
			Method:  method,
			Target:  target,
			Version: version,
		},
		remaining,
	}, nil
}

func ParseHeaders(ps *ParseState) (*ParseState, error) {
	withoutBody, remaining, found := bytes.Cut(ps.Remaining, CRLF2)

	if !found {
		return nil, errors.New("malformed http message")
	}

	_, rawHeaders, found := bytes.Cut(withoutBody, CRLF)

	if !found {
		return nil, errors.New("malformed http request line or headers")
	}

	headers := make(map[string][]byte)

	for v := range bytes.SplitSeq(rawHeaders, CRLF) {
		headerName, headerValue, found := bytes.Cut(v, []byte(":"))

		if !found {
			return nil, errors.New("malformed headers")
		}

		if !utf8.Valid(headerName) {
			return nil, errors.New("header name is not valid utf-8")
		}

		headers[string(bytes.ToLower(headerName))] = bytes.TrimSpace(headerValue)
	}

	ps.Msg.Headers = headers
	ps.Remaining = remaining

	return ps, nil
}

func ParseBody(ps *ParseState) (*ParseState, error) {
	contentLength, err := strconv.Atoi(string(ps.Msg.Headers["content-length"]))

	if err != nil {
		return nil, err
	}

	if contentLength <= 0 {
		return ps, nil
	}

	if len(ps.Remaining) < contentLength {
		return nil, ErrBodyTooSmall
	}

	ps.Msg.Body = ps.Remaining[:contentLength]

	return ps, nil
}

func ParseRequest(s []byte) (HttpRequestMessage, error) {
	ps, err := ParseRequestLine(s)

	if err != nil {
		return HttpRequestMessage{}, err
	}

	psWithHeaders, err := ParseHeaders(&ps)

	if err != nil {
		return HttpRequestMessage{}, err
	}

	if psWithHeaders.Msg.Headers["content-length"] == nil {
		return psWithHeaders.Msg, nil
	}

	psWithBody, err := ParseBody(psWithHeaders)

	if err != nil {
		return HttpRequestMessage{}, err
	}

	return psWithBody.Msg, nil
}

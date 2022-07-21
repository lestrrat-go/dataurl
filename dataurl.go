//go:generate ./tools/cmd/genoptions.sh

package dataurl

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

// MediaType holds the relevant type information for the data.
type MediaType struct {
	Type   string
	Params map[string]string
}

// URL represents a data URL structure.
type URL struct {
	MediaType MediaType
	Data      []byte
}

var scheme = []byte(`data:`)
var base64Marker = []byte(`;base64`)
var b64enc = base64.StdEncoding

func defaultMediaType() MediaType {
	return MediaType{
		Type: `text/plain`,
		Params: map[string]string{
			`charset`: `US-ASCII`,
		},
	}
}

// Parse takes a data URL as a sequence of bytes and parses it into
// `*dataurl.URL` object.
//
func Parse(data []byte) (*URL, error) {
	if !bytes.HasPrefix(data, scheme) {
		return nil, fmt.Errorf(`invalid scheme`)
	}
	data = data[len(scheme):]

	switch {
	case len(data) < 1:
		return nil, fmt.Errorf(`invalid data URL (no data)`)
	case data[0] == ',': // data:,xxxxx
		return parseData(defaultMediaType(), false, data)
	case data[0] == ';': // data:;base64,xxxxx
		return parseBase64Marker(defaultMediaType(), data)
	default:
		// data:.....;foo=bar;base64....
		i := bytes.Index(data, base64Marker)
		var parseBase64 bool
		if i > 0 {
			parseBase64 = true
		} else {
			i = bytes.IndexByte(data, ',')
		}

		if i < 0 {
			return nil, fmt.Errorf(`invalid data URL (no data)`)
		}

		// The parsing logic for media type parameters is curretly completely
		// defered to "mime.ParseMEdiaType()". If it can parse it, then we think
		// it's valid -- but we _DO_ unescape attribute keys anr values
		// if they contain % signs. I don't know, it looks weird, but we'll go with this for now

		typ, params, err := mime.ParseMediaType(string(data[:i]))
		if err != nil {
			return nil, fmt.Errorf(`failed to parse media type %q`, data[:i])
		}

		for k, v := range params {
			unescapedKey, err := unescape([]byte(k), false)
			if err != nil {
				return nil, fmt.Errorf(`failed to unescape parameter key %q: %w`, k, err)
			}

			unescapedValue, err := unescape([]byte(v), false)
			if err != nil {
				return nil, fmt.Errorf(`failed to unescape parameter value for %q: %w`, k, err)
			}

			if uks := string(unescapedKey); uks != k {
				delete(params, k)
				params[uks] = string(unescapedValue)
				continue
			}

			if uvs := string(unescapedValue); v != uvs {
				params[k] = uvs
			}
		}

		data = data[i:]

		mt := MediaType{
			Type:   typ,
			Params: params,
		}

		if parseBase64 {
			return parseBase64Marker(mt, data)
		} else {
			return parseData(mt, false, data)
		}
	}
}

func parseBase64Marker(mediaType MediaType, data []byte) (*URL, error) {
	if !bytes.HasPrefix(data, base64Marker) {
		return nil, fmt.Errorf(`invalid data URL (invalid base64 marker)`)
	}
	return parseData(mediaType, true, data[len(base64Marker):])
}

func parseData(mediaType MediaType, isBase64 bool, data []byte) (*URL, error) {
	if len(data) < 2 || data[0] != ',' {
		return nil, fmt.Errorf(`invalid data URL (invalid data section)`)
	}
	data = data[1:]

	ret := URL{
		MediaType: mediaType,
	}
	if !isBase64 {
		unescaped, err := unescape(data, true)
		if err != nil {
			return nil, fmt.Errorf(`invalid data URL (failed to escape data: %w)`, err)
		}
		ret.Data = unescaped
	} else {
		dst := make([]byte, b64enc.DecodedLen(len(data)))
		n, err := b64enc.Decode(dst, data)
		if err != nil {
			return nil, fmt.Errorf(`invalid data URL (base64: %w, %q)`, err, data)
		}
		ret.Data = dst[:n]
	}
	return &ret, nil
}

func unescape(data []byte, strict bool) ([]byte, error) {
	var base [1]byte
	var dst bytes.Buffer
	var buf = base[:1]

	l := len(data)
	max := l - 1
	dst.Grow(l) // even with no escape, it's going to be _around_ the same size as the origianl data
	for i := 0; i < l; i++ {
		switch c := data[i]; c {
		case '%':
			if i > max-2 { // need two more bytes
				return nil, fmt.Errorf(`failed to unescape: unexpected end of byte sequence at byte %d`, i)
			}

			if _, err := hex.Decode(buf, data[i+1:i+3]); err != nil {
				return nil, fmt.Errorf(`failed to unescape: invalid hexadecimal sequence starting at byte %d`, i)
			}
			dst.Write(buf)
			i += 2
		default:
			if !strict {
				dst.WriteByte(c)
				continue
			}

			if isNotReserved(c) {
				dst.WriteByte(c)
				continue
			}

			return nil, fmt.Errorf(`failed to unescape: reserved character %q found at byte %d`, c, i)
		}
	}
	return dst.Bytes(), nil
}

// Encode encodes a piece of data into data URL format.
//
// By default this function auto-detects the content of the given piece of
// data using `"net/http".DetectContentType`.
//
// Users may override this by passing an explicit media type by using
// a combination of `dataurl.WithMediaType()` and `url.WithMediaTypeParams()` options.
//
// Also by default the data is encoded using base64 encoding when
// the media type is anything other than a `text/****` type.
//
// You may override this by using the `dataurl.WithBase64Encoding()` option.
func Encode(data []byte, options ...EncodeOption) ([]byte, error) {
	var dst bytes.Buffer
	var mt string
	var params map[string]string
	var explicitBase64 bool // true if the user specified base64
	var encodeBase64 bool
	for _, option := range options {
		switch option.Ident() {
		case identMediaType{}:
			mt = option.Value().(string)
		case identMediaTypeParams{}:
			params = option.Value().(map[string]string)
		case identBase64Encoding{}:
			explicitBase64 = true
			encodeBase64 = option.Value().(bool)
		}
	}

	if mt == "" {
		mt = http.DetectContentType(data)
	}

	if strings.IndexByte(mt, ';') > -1 {
		if len(params) != 0 {
			// the user specified a media type with parameters, _AND_
			// gave us more parameters to work with
			parsedMt, parsedParams, err := mime.ParseMediaType(mt)
			if err != nil {
				return nil, fmt.Errorf(`failed to parse media type: %w`, err)
			}

			// merge parsedParams and params. params takes precedence,
			// so overwrite it
			for k, v := range params {
				parsedParams[k] = v
			}
			mt = mime.FormatMediaType(parsedMt, parsedParams)
		}
	} else if len(params) != 0 {
		// mt is something like 'text/plain', and we have extra parameters
		mt = mime.FormatMediaType(mt, params)
	}

	// It is possible that either the user or the library that we depend
	// on provides us with a media type that is tiny bit off from what
	// we want...
	//
	// namely: https://cs.opensource.google/go/go/+/refs/tags/go1.18.4:src/net/http/sniff.go;l=308
	//
	// Here, DetectContentType may return a media type with a space after the semicolon,
	// which is not good for our case. Forcefully fix it
	mt = strings.Replace(mt, `; `, `;`, -1)

	dst.Write(scheme)
	dst.WriteString(mt)

	if !explicitBase64 {
		// The user has not explicitly provided us with the option to
		// either use or not use base64. We're going to use base64
		// if and only if the data is not a text-type
		if !strings.HasPrefix(mt, `text`) {
			// use base64
			encodeBase64 = true
		}
	}

	if encodeBase64 {
		dst.Write(base64Marker)
	}

	dst.WriteByte(',')

	if encodeBase64 {
		enc := base64.NewEncoder(b64enc, &dst)
		_, _ = enc.Write(data)
		enc.Close()
	} else {
		writeEscapedSequence(&dst, data)
	}

	return dst.Bytes(), nil
}

func isNotReserved(b byte) bool {
	return (b >= '0' && b <= '9') || // 0-9
		(b >= 'a' && b <= 'z') || // a-z
		(b >= 'A' && b <= 'Z') || // A-Z
		(b >= '\'' && b <= '*') || // ', (, ), *
		(b >= '-' && b <= '.') || // -, .
		b == '!' ||
		b == '_' ||
		b == '~'
}

func writeEscapedSequence(dst *bytes.Buffer, data []byte) {
	for _, b := range data {
		if isNotReserved(b) {
			dst.WriteByte(b)
			continue
		}
		fmt.Fprintf(dst, `%%%02X`, b)

	}
}

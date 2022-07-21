// This file is auto-generated by tools/cmd/genoptions/main.go. DO NOT EDIT

package dataurl

import (
	"github.com/lestrrat-go/option"
)

type Option = option.Interface

// EncodeOption is a type of option that can be passed to Encode()
type EncodeOption interface {
	Option
	encodeOption()
}

type encodeOption struct {
	Option
}

func (*encodeOption) encodeOption() {}

type identBase64Encoding struct{}
type identMediaType struct{}
type identMediaTypeParams struct{}

func (identBase64Encoding) String() string {
	return "WithBase64Encoding"
}

func (identMediaType) String() string {
	return "WithMediaType"
}

func (identMediaTypeParams) String() string {
	return "WithMediaTypeParams"
}

// WithBase64Encoding specifies if the payload should or should not
// be base64 encoded. Specifying this option overrides the automatic
// detection that is performed by default, where any payload without
// an explciit `text/****` media type will be base64 encoded
func WithBase64Encoding(v bool) EncodeOption {
	return &encodeOption{option.New(identBase64Encoding{}, v)}
}

// WithMediaType allows users to specify an explciit media type for the
// data to be encoded.
//
// If unspecified, `"net/http".DetectContentType` will be used to sniff
// the media type.
//
// You may include parameters (e.g. `charset=utf-8`) in this string as well,
// but it is the caller's responsibility to make sure that it is well-formed.
func WithMediaType(v string) EncodeOption {
	return &encodeOption{option.New(identMediaType{}, v)}
}

// WithMediaTypeParams allows users to provide extra media type parameters,
// such as `charset=utf-8` as a map.
//
// Upon any conflict, values provided in this map will overwrite the
// values found in the media type string provided either explcitly by the
// user or by auto-detection.
//
// It is the user's reponsibility to properly format the parameter names,
// such as properly making everything lower-case (or not).
func WithMediaTypeParams(v map[string]string) EncodeOption {
	return &encodeOption{option.New(identMediaTypeParams{}, v)}
}
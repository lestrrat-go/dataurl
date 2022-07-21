package dataurl_test

import (
	"encoding/base64"
	"testing"

	"github.com/lestrrat-go/dataurl"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testcases := []struct {
		Name     string
		Data     []byte
		Error    bool
		Expected *dataurl.URL
	}{
		{
			Name: `sample from RFC`,
			Data: []byte(`data:image/gif;base64,R0lGODdhMAAwAPAAAAAAAP///ywAAAAAMAAwAAAC8IyPqcvt3wCcDkiLc7C0qwyGHhSWpjQu5yqmCYsapyuvUUlvONmOZtfzgFzByTB10QgxOR0TqBQejhRNzOfkVJ+5YiUqrXF5Y5lKh/DeuNcP5yLWGsEbtLiOSpa/TPg7JpJHxyendzWTBfX0cxOnKPjgBzi4diinWGdkF8kjdfnycQZXZeYGejmJlZeGl9i2icVqaNVailT6F5iJ90m6mvuTS4OK05M0vDk0Q4XUtwvKOzrcd3iq9uisF81M1OIcR7lEewwcLp7tuNNkM3uNna3F2JQFo97Vriy/Xl4/f1cf5VWzXyym7PHhhx4dbgYKAAA7`),
			Expected: &dataurl.URL{
				MediaType: dataurl.MediaType{
					Type:   `image/gif`,
					Params: map[string]string{},
				},
				Data: (func(data string) []byte {
					v, err := base64.StdEncoding.DecodeString(data)
					if err != nil {
						panic(err)
					}
					return v
				})(`R0lGODdhMAAwAPAAAAAAAP///ywAAAAAMAAwAAAC8IyPqcvt3wCcDkiLc7C0qwyGHhSWpjQu5yqmCYsapyuvUUlvONmOZtfzgFzByTB10QgxOR0TqBQejhRNzOfkVJ+5YiUqrXF5Y5lKh/DeuNcP5yLWGsEbtLiOSpa/TPg7JpJHxyendzWTBfX0cxOnKPjgBzi4diinWGdkF8kjdfnycQZXZeYGejmJlZeGl9i2icVqaNVailT6F5iJ90m6mvuTS4OK05M0vDk0Q4XUtwvKOzrcd3iq9uisF81M1OIcR7lEewwcLp7tuNNkM3uNna3F2JQFo97Vriy/Xl4/f1cf5VWzXyym7PHhhx4dbgYKAAA7`),
			},
		},
		{
			Name: `skip media type`,
			Data: []byte(`data:,hello%2C%20world!`),
			Expected: &dataurl.URL{
				MediaType: dataurl.MediaType{
					Type: `text/plain`,
					Params: map[string]string{
						`charset`: `US-ASCII`,
					},
				},
				Data: []byte(`hello, world!`),
			},
		},
		{
			Name: `skip media type, use base64`,
			Data: []byte(`data:;base64,aGVsbG8sIHdvcmxkIQ==`),
			Expected: &dataurl.URL{
				MediaType: dataurl.MediaType{
					Type: `text/plain`,
					Params: map[string]string{
						`charset`: `US-ASCII`,
					},
				},
				Data: []byte(`hello, world!`),
			},
		},
		{
			Name: `odd values`,
			Data: []byte(`data:application/json;charset=utf-8;oddParam1="a\"<@>\"z";odd%20param2=hello%20world;base64,eyJoZWxsbyI6IndvcmxkIn0=`),
			Expected: &dataurl.URL{
				MediaType: dataurl.MediaType{
					Type: `application/json`,
					Params: map[string]string{
						`charset`:    `utf-8`,
						`oddparam1`:  `a"<@>"z`,
						`odd param2`: `hello world`,
					},
				},
				Data: []byte(`{"hello":"world"}`),
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			u, err := dataurl.Parse(tc.Data)
			if tc.Error {
				require.Error(t, err, `dataurl.Parse should fail`)
				return
			}

			require.NoError(t, err, `dataurl.Parse should succeed`)
			require.Equal(t, tc.Expected, u)
		})
	}
}

func TestEncode(t *testing.T) {
	testcases := []struct {
		Data     []byte
		Expected []byte
		Options  []dataurl.EncodeOption
		Error    bool
	}{
		{
			Data:     []byte(`hello, world!`),
			Expected: []byte(`data:text/plain;charset=utf-8,hello%2C%20world!`),
		},
		{
			Data: []byte(`{"hello":"world"}`),
			Options: []dataurl.EncodeOption{
				dataurl.WithMediaType(`application/json`),
				dataurl.WithMediaTypeParams(map[string]string{
					`charset`: `utf-8`,
				}),
			},
			Expected: []byte(`data:application/json;charset=utf-8;base64,eyJoZWxsbyI6IndvcmxkIn0=`),
		},
		{
			Data: []byte(`{"hello":"world"}`),
			Options: []dataurl.EncodeOption{
				dataurl.WithMediaType(`application/json; charset=US-ASCII`),
				dataurl.WithMediaTypeParams(map[string]string{
					`charset`: `utf-8`,
				}),
			},
			Expected: []byte(`data:application/json;charset=utf-8;base64,eyJoZWxsbyI6IndvcmxkIn0=`),
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(string(tc.Data), func(t *testing.T) {
			u, err := dataurl.Encode(tc.Data, tc.Options...)
			if tc.Error {
				require.NoError(t, err, `dataurl.Encode should fail`)
				return
			}
			require.NoError(t, err, `dataurl.Encode should succeed`)
			require.Equal(t, tc.Expected, u, `results should match`)

			parsed, err := dataurl.Parse(u)
			require.NoError(t, err, `dataurl.Parse should succeed`)
			require.Equal(t, tc.Data, parsed.Data, `data should match`)
		})
	}

}

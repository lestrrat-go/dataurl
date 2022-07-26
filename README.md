dataurl
=======

Parse and encode RFC2397 "data' URL scheme.

# SYNOPSIS

## Encoding

<!-- INCLUDE(examples/encode_example_test.go) -->
```go
package examples

import (
  "encoding/base64"
  "fmt"

  "github.com/lestrrat-go/dataurl"
)

func ExampleEncode() {
  { // Let dataurl.Encode fiture out the media type
    encoded, err := dataurl.Encode([]byte(`Hello, World!`))
    if err != nil {
      fmt.Printf("failed to encode: %s", err)
      return
    }
    const expected = `data:text/plain;charset=utf-8,Hello%2C%20World!`
    if string(encoded) != expected {
      fmt.Printf("expected: %q\n", expected)
      fmt.Printf("actual:   %q\n", encoded)
      return
    }
  }

  { // It works on binary files too
    rawimg, err := base64.StdEncoding.DecodeString(gopher)
    if err != nil {
      fmt.Printf("failed to decode gopher image: %s", err)
      return
    }

    encoded, err := dataurl.Encode(rawimg)
    if err != nil {
      fmt.Printf("failed to encode: %s", err)
      return
    }

    var expected = `data:image/png;base64,` + gopher
    if string(encoded) != expected {
      fmt.Printf("expected: %q\n", expected)
      fmt.Printf("actual:   %q\n", encoded)
      return
    }
  }

  { // You can overwrite the media type

    { // First case, only supply the main media type
      encoded, err := dataurl.Encode(
        []byte(`{"Hello":"World!"}`),
        dataurl.WithMediaType(`application/json`),
      )

      if err != nil {
        fmt.Printf("failed to encode: %s", err)
        return
      }
      const expected = `data:application/json;base64,eyJIZWxsbyI6IldvcmxkISJ9`
      if string(encoded) != expected {
        fmt.Printf("expected: %q\n", expected)
        fmt.Printf("actual:   %q\n", encoded)
        return
      }
    }

    { // Second case, supply the mdia type and parameters as a string
      encoded, err := dataurl.Encode(
        []byte(`{"Hello":"World!"}`),
        dataurl.WithMediaType(`application/json; charset=utf-8`),
      )

      if err != nil {
        fmt.Printf("failed to encode: %s", err)
        return
      }
      const expected = `data:application/json;charset=utf-8;base64,eyJIZWxsbyI6IldvcmxkISJ9`
      if string(encoded) != expected {
        fmt.Printf("expected: %q\n", expected)
        fmt.Printf("actual:   %q\n", encoded)
        return
      }
    }

    { // Third case, supply the mdia type as string, and parameters as a map
      // Notice that the parameters OVERWRITE the value given in the string media type
      encoded, err := dataurl.Encode(
        []byte(`{"Hello":"World!"}`),
        dataurl.WithMediaType(`application/json; charset=us-ascii`),
        dataurl.WithMediaTypeParams(map[string]string{
          `charset`: `utf-8`,
        }),
      )

      if err != nil {
        fmt.Printf("failed to encode: %s", err)
        return
      }
      const expected = `data:application/json;charset=utf-8;base64,eyJIZWxsbyI6IldvcmxkISJ9`
      if string(encoded) != expected {
        fmt.Printf("expected: %q\n", expected)
        fmt.Printf("actual:   %q\n", encoded)
        return
      }
    }
  }

  { // Explicitly specify to enable or disable base64

    { // First case: by default this would NOT be base64 encoded,
      // but you can force it to do so
      encoded, err := dataurl.Encode(
        []byte(`Hello, World!`),
        dataurl.WithBase64Encoding(true),
      )
      if err != nil {
        fmt.Printf("failed to encode: %s", err)
        return
      }
      const expected = `data:text/plain;charset=utf-8;base64,SGVsbG8sIFdvcmxkIQ==`
      if string(encoded) != expected {
        fmt.Printf("expected: %q\n", expected)
        fmt.Printf("actual:   %q\n", encoded)
        return
      }
    }

    { // Second case: by defualt his would be base64 encoded,
      // but you can force it to emit plain text
      // Note that this would produce really bad results if your data is
      // actually binary
      encoded, err := dataurl.Encode(
        []byte(`{"Hello":"World!"}`),
        dataurl.WithMediaType(`application/json`),
        dataurl.WithBase64Encoding(false),
      )

      if err != nil {
        fmt.Printf("failed to encode: %s", err)
        return
      }
      const expected = `data:application/json,%7B%22Hello%22%3A%22World!%22%7D`
      if string(encoded) != expected {
        fmt.Printf("expected: %q\n", expected)
        fmt.Printf("actual:   %q\n", encoded)
        return
      }
    }
  }

  // OUTPUT:
  //
}

const gopher = `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AmA0oKDz1CJYbzAYjE63BHUKhXLp0KHDGlevXk0aGhrIqVOnyLvvvksefvhh8q9//YtUVVVdVfJyQUEB2bp1K/nTn/5EIiOjzFqtx0EAf5TVTklHdnt3vRZjMBjdBR04TqcipGa2Xq9/PClp+iiO43STJk2Ch4cHDhw4gLq6OsyaNQtTpkyBWn3xjBSHw4GCgnwcPHgIJ0+exJ49e8y5uXkZVqslHeD2A0IK4KgDBgLI6bqPDAaD0X1QKqPoSmoqX45zvP/uuyseWrRoEQBAFEUcOXIEn3zyCcLDw/H444/DarWiqqoKRqMRVqsVRqMRpaVlUnFxcXVGxsmcEyeO14qiPRPgfgUUx5TKIbWimE5oef4uDcsDZDC6GbL4aQBY+YkT7wpNSkpyE0clxo4dC6PRiMcff9y4a9fuXJPJxNXU1JSZzWa1KIqVDodY6XCIWYSQUwDOAp5mIEoEmgCUQhTTWCMzGIyb1QKMRN++3lCptA++//775vPjeY2NjeSxxx6TAGEZwPUB0AfgPAHBG/BUAr4AfFhDMhiMrogGAPrNmDHzl5KSkgtGcN944w2Th4d+LaAIBpSsuRgMRvdAELwBgPf1DViybt1n0vnW35EjR0hkZNQGAB6sIDxrAQajG4mfDyTJAgAht94akxwZGcHt3Lmz+X2Hw4Hvv//Bkp9ftMPDw9PcHQYxmAAyGAwAACEEAAEAtVqt0jkcDmzYsAEVFfJ64pmZmdi8eUsmIeJPFosElSqYPTTYZcNgdA/0+hkQxUIQYrdYrY6+gYEBQwsLC5WnT59Gfn4+li1b5jh5MuujESOcu6qq+kAUy3p8m7E8QAajWzp2kk9gYNA0QsgYg8GglCRHjN1uO8bzutckyVxNbcYe31L/D3E/BrXKhWzDAAAARnRFWHRjb21tZW50AEZpbGUgc291cmNlOiBodHRwOi8vY29tbW9ucy53aWtpbWVkaWEub3JnL3dpa2kvRmlsZTpHb2xhbmcucG5nqG181wAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxNC0wOS0xNlQwMzo1NzowNCswMDowMPaiDFMAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTQtMDktMTZUMDM6NTc6MDQrMDA6MDCH/7TvAAAARnRFWHRzb2Z0d2FyZQBJbWFnZU1hZ2ljayA2LjYuOS03IDIwMTQtMDMtMDYgUTE2IGh0dHA6Ly93d3cuaW1hZ2VtYWdpY2sub3JngdOzwwAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpoZWlnaHQANDQwUmuvDwAAABh0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAAxMjI08evQygAAABl0RVh0VGh1bWI6Ok1pbWV0eXBlAGltYWdlL3BuZz+yVk4AAAAXdEVYdFRodW1iOjpNVGltZQAxNDEwODM5ODI0DeAXbwAAABN0RVh0VGh1bWI6OlNpemUAODAuMUtCQr9J6A8AAAAzdEVYdFRodW1iOjpVUkkAZmlsZTovLy90bXAvbG9jYWxjb3B5XzQ2YjJiMTRkMDQwOS0xLnBuZ8kGguYAAAAASUVORK5CYII=`
```
source: [examples/encode_example_test.go](https://github.com/lestrrat-go/dataurl/blob/main/examples/encode_example_test.go)
<!-- END INCLUDE -->

## Parsing

<!-- INCLUDE(examples/parse_example_test.go) -->
```go
package examples

import (
  "fmt"

  "github.com/lestrrat-go/dataurl"
)

func ExampleParse() {
  u, err := dataurl.Parse([]byte(`data:application/json;charset=utf-8;base64,eyJIZWxsbyI6IldvcmxkISJ9`))
  if err != nil {
    fmt.Printf("failed to parse: %s", err)
    return
  }

  fmt.Printf("media type: %q\n", u.MediaType.Type)
  fmt.Printf("params:\n")
  for k, v := range u.MediaType.Params {
    fmt.Printf("  %s: %s\n", k, v)
  }
  fmt.Printf("data: %s\n", u.Data)

  // OUTPUT:
  // media type: "application/json"
  // params:
  //   charset: utf-8
  // data: {"Hello":"World!"}
}
```
source: [examples/parse_example_test.go](https://github.com/lestrrat-go/dataurl/blob/main/examples/parse_example_test.go)
<!-- END INCLUDE -->

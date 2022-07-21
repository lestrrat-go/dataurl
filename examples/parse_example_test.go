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

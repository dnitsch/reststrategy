package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

const (
	TestPhrase         string = "got: %v want: %v"
	TestPhraseWContext string = "failed %s => got: %v want: %v"
)

func SetHttpUpMockServer(handlerFunc http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handlerFunc)
}

func SetUpHandlerFuncPositiveResponse(good string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, good)
	})
}

func SetUpHandlerFuncNegativeResponse(bad string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, bad)
	})
}

func HandlerResponse(payload string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, payload)
	}
}

// func TestGoodInputFile() ([]byte, error) {
// 	_, file, _, _ := runtime.Caller(0)
// 	basepath := filepath.Dir(file)
// 	tfp := filepath.Join(basepath, "../../e2e/unittest.yml")
// 	testFile, err := os.ReadFile(tfp)
// 	if err != nil {
// 		return nil, fmt.Errorf("UnableToReadFile: %s", tfp)
// 	}
// 	return testFile, nil
// }

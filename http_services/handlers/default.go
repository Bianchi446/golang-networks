package handlers

import(
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
)

var t = template.Must(template.New("Hello").Parse("Hello", {{.}}! ))

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r, *http.Request) {
			defer func(r io.ReadCloser) {
				_, _ = io.Copy(ioutil.Discard, r)
				_ = r.Close()
			} (r.Body)

			var b []byte

			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				b, err = ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "internal server error", 
						http.StatusInternalServerError)
					return 
				}
			default:
				// Not RFC - compliant 
				http.Error(w, "Method not allowed", 
					http.StatusMethodNotAllowed)
				return 
			}

			_ = t.Execute(w, string(b))
		},
	)
}
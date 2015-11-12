package irondog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type MarkdownPostParam struct {
	Text    string `json:"text"`
	Mode    string `json:"mode"`
	Context string `json:"context"`
}

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	fp, err := os.Open("md/sample.md")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fp.Close()

	b, err := ioutil.ReadAll(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	param := MarkdownPostParam{
		Text:    string(b),
		Mode:    "gfm",
		Context: "github/gollum",
	}
	paramBytes, err := json.Marshal(param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := urlfetch.Client(ctx)
	resp, err := client.Post("https://api.github.com/markdown", "application/json", bytes.NewReader(paramBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Errorf(ctx, "reponse write error = %v", err)
	}
}

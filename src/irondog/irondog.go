package irondog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"

	"google.golang.org/appengine"
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

	md, err := readMarkdownFile(r.URL.Path[1:len(r.URL.Path)])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	param := MarkdownPostParam{
		Text:    string(md),
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

	header, err := readHtmlFile("header")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	footer, err := readHtmlFile("footer")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Header  string
		Content string
		Footer  string
	}{
		Header:  string(header),
		Content: string(body),
		Footer:  string(footer),
	}

	mainTempl, err := template.ParseFiles("html/main.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = mainTempl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func readHtmlFile(path string) ([]byte, error) {
	fp, err := os.Open(fmt.Sprintf("html/%s.html", path))
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return ioutil.ReadAll(fp)
}

func readMarkdownFile(path string) ([]byte, error) {
	fp, err := os.Open(fmt.Sprintf("md/%s.md", path))
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return ioutil.ReadAll(fp)
}

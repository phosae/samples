package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var apifileRegEx *regexp.Regexp

func init() {
	var err error
	apifileRegEx, err = regexp.Compile(`^.+\.(yaml|json)$`)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	filedir := os.Getenv("STATIC_FILE_DIR")
	if filedir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			filedir = home
		} else {
			filedir = "."
		}
	}

	if args := os.Args; len(args) > 1 {
		if URL, err := url.Parse(args[len(args)-1]); err == nil {
			if resp, err := http.Get(URL.String()); err == nil && resp.StatusCode/100 == 2 {
				defer resp.Body.Close()
				parts := strings.Split(URL.Path, "/")
				name := parts[len(parts)-1]
				if b, err := io.ReadAll(resp.Body); err == nil {
					os.WriteFile(filepath.Join(filedir, "spec", name), b, 0666)
				}
			}
		}
	}

	http.Handle("/", http.FileServer(http.Dir(filedir)))
	log.Printf("serve fs %s\n", filedir)

	http.Handle("/apifiles", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var files []string
		err := filepath.Walk(filepath.Join(filedir, "spec"), func(_ string, info os.FileInfo, err error) error {
			if err == nil && apifileRegEx.MatchString(info.Name()) {
				files = append(files, info.Name())
			}
			return nil
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		data, err := json.Marshal(files)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(data)
	}))

	http.Handle("/uis", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var uis = []string{"elements", "rapidoc", "swagger", "redoc"}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		data, err := json.Marshal(uis)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(data)
	}))

	http.Handle("/view", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
			w.Header().Set("Access-Control-Max-Age", "86400")
			return
		}

		type ViewReq struct {
			UI      string `json:"ui"`
			ApiFile string `json:"apifile"`
		}
		var viewReq ViewReq

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&viewReq)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		viewURL := fmt.Sprintf("/%s/%s", viewReq.UI, viewReq.ApiFile)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		http.Redirect(w, r, viewURL, http.StatusFound)
	}))

	// https://api.apis.guru/v2/specs/github.com/1.1.4/openapi.yaml
	http.Handle("/elements/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/elements/")
		spec := parts[len(parts)-1]
		spec = fmt.Sprintf("/spec/%s", spec)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Write([]byte(fmt.Sprintf(Elements, spec)))
	}))

	http.Handle("/rapidoc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/rapidoc/")
		spec := parts[len(parts)-1]
		spec = fmt.Sprintf("/spec/%s", spec)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Write([]byte(fmt.Sprintf(RapiDoc, spec)))
	}))

	http.Handle("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/swagger/")
		spec := parts[len(parts)-1]
		spec = fmt.Sprintf("/spec/%s", spec)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Write([]byte(fmt.Sprintf(SwaggerUI, spec)))
	}))

	http.Handle("/redoc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/redoc/")
		spec := parts[len(parts)-1]
		spec = fmt.Sprintf("/spec/%s", spec)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Write([]byte(fmt.Sprintf(Redoc, spec)))
	}))

	log.Fatalln(http.ListenAndServe(":8000", http.DefaultServeMux))
}

// inspired by https://www.jvt.me/posts/2022/03/16/kubernetes-openapi/
// https://github.com/stoplightio/elements
var Elements = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Elements in HTML</title>
    <!-- Embed elements Elements via Web Component -->
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>

    <elements-api
      apiDescriptionUrl="%s"
      router="hash"
      layout="sidebar"
    />

  </body>
</html>`

// https://github.com/rapi-doc/RapiDoc
var RapiDoc = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8"> <!-- Important: rapi-doc uses utf8 characters -->
    <script type="module" src="https://unpkg.com/rapidoc/dist/rapidoc-min.js"></script>
  </head>
  <body>
    <rapi-doc spec-url = "%s"> </rapi-doc>
  </body>
</html>`

// https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/installation.md
var SwaggerUI = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta
      name="description"
      content="SwaggerUI"
    />
    <title>SwaggerUI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
  </head>
  <body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '%s',
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
      });
    };
  </script>
  </body>
</html>`

// https://github.com/Redocly/redoc
var Redoc = `<!DOCTYPE html>
<html>
  <head>
    <title>Redoc</title>
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <!--
    Redoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='%s'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
  </body>
</html>`

// more tools https://openapi.tools/#documentation

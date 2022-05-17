/*
Copyright © 2021-2022 Infinite Devices GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package router

import (
	"encoding/json"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	"github.com/infinimesh/http-fs/pkg/io"
)

// Creates new mux Router with default routes settings
// Default routes:
// 	GET /{ns} - returns stats (files and their props) in current namespace(dir)
// 	DELETE /{ns} - deletes namespace(dir)
// 	GET /{ns}/{file} - returns file itself
// 	POST /{ns}/{file} - upload file
// 	DELETE /{ns}/{file} - deletes file
func NewRouter(h io.IOHandler) *mux.Router {
	r := mux.NewRouter()
	
	r.HandleFunc("/{ns}", Stat(h)).Methods("GET")
	r.HandleFunc("/{ns}/{file}", Fetch(h)).Methods("GET")

	r.Use(mux.CORSMethodMiddleware(r))

	return r
}

func Stat(h io.IOHandler) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := mux.Vars(r)["ns"]
		files, err := h.Stat(ns)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(files)
	}
}

func Fetch(h io.IOHandler) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		ns := mux.Vars(r)["ns"]
		file := mux.Vars(r)["file"]
		f, mime, err := h.Fetch(ns, file)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if mime == nil {
			m := mimetype.Detect(f).String()
			mime = &m
		}
		w.Header().Set("Content-Type", *mime)
		w.Write(f)
	}
}

		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}
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
	"io/ioutil"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	"github.com/infinimesh/http-fs/pkg/io"
	"github.com/infinimesh/http-fs/pkg/mw"
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
	r.HandleFunc("/{ns}", Delete(h)).Methods("DELETE")

	r.HandleFunc("/{ns}/{file}", Fetch(h)).Methods("GET")
	r.HandleFunc("/{ns}/{file}", Upload(h)).Methods("POST")
	r.HandleFunc("/{ns}/{file}", Delete(h)).Methods("DELETE")

	r.Use(mux.CORSMethodMiddleware(r))

	return r
}

func Access(r *http.Request) (read, write bool) {
	v := r.Context().Value(mw.AccessKey)
	if v == nil {
		return false, false
	}

	access, ok := v.(mw.Access)
	if !ok {
		return false, false
	}

	return access.Read, access.Write
}

func Stat(h io.IOHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		read, _ := Access(r)
		if !read {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ns := mux.Vars(r)["ns"]
		files, err := h.Stat(ns)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(files)
	}
}

func Fetch(h io.IOHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		read, _ := Access(r)
		if !read {
			w.WriteHeader(http.StatusForbidden)
			return
		}

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

func Upload(h io.IOHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, write := Access(r)
		if !write {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ns := mux.Vars(r)["ns"]
		filename := mux.Vars(r)["file"]

		if ns == "" || filename == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Namespace and Filename are required"))
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: File must be present as form under key 'file'"))
			return
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error: Failed to read file"))
			return
		}

		err = h.Upload(ns, filename, bytes)
		if err != nil {
			if err, ok := err.(*io.FileTooLargeError); ok {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(err.Error()))
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func Delete(h io.IOHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, write := Access(r)
		if !write {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ns := mux.Vars(r)["ns"]
		file, ok := mux.Vars(r)["file"]

		var err error
		if ok {
			err = h.Delete(ns, file)
		} else {
			err = h.DeleteNS(ns)
		}

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

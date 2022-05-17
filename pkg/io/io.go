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
package io

type IOHandler interface {
	// returns stats (files and their props) in current namespace(dir)
	Stat(ns string) ([]File, error)
	// returns file itself and optionally mime type
	Fetch(ns, file string) (bytes []byte, mime *string, err error)
	// writes file
	Upload(ns, file string, bytes []byte) error

	// deletes file
	Delete(ns, file string) error
	// deletes namespace(dir)
	DeleteNS(ns string) error
}

type File struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

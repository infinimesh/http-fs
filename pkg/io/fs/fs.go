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
package fs

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/infinimesh/http-fs/pkg/io"
	"go.uber.org/zap"
)

type FSHandler struct {
	log  *zap.Logger
	root string
}

func NewFileSystemHandler(log *zap.Logger, root string) *FSHandler {
	return &FSHandler{
		log:  log.Named("fs"),
		root: root,
	}
}

func (f *FSHandler) Stat(ns string) ([]io.File, error) {
	log := f.log.Named("Stat")

	p := path.Join(f.root, ns)
begin:
	files, err := ioutil.ReadDir(p)
	if os.IsNotExist(err) {
		log.Warn("Namespace does not exist", zap.String("ns", ns))
		err = os.Mkdir(p, 0755)
		if err != nil {
			panic(err)
		}
		goto begin
	}
	if err != nil {
		log.Error("failed to read directory", zap.Error(err))
		return nil, err
	}

	content := make([]io.File, 0)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		content = append(content, io.File{Name: f.Name(), Size: f.Size()})
	}

	return content, nil
}

func (f *FSHandler) Fetch(ns, file string) ([]byte, *string, error) {
	log := f.log.Named("Fetch")

	p := path.Join(f.root, ns, file)
	bytes, err := os.ReadFile(p)

	if err != nil {
		log.Error("failed to read file", zap.Error(err))
		return nil, nil, err
	}

	return bytes, nil, nil
}

func (f *FSHandler) Upload(ns, file string, data []byte) error {
	log := f.log.Named("Upload")

	p := path.Join(f.root, ns, file)

begin:
	err := os.WriteFile(p, data, 0644)
	if os.IsNotExist(err) {
		log.Warn("Namespace does not exist", zap.String("ns", ns))
		err = os.Mkdir(path.Join(f.root, ns), 0755)
		if err != nil {
			panic(err)
		}
		goto begin
	}
	if err != nil {
		log.Error("failed to write file", zap.Error(err))
		return err
	}
	return nil
}

func (f *FSHandler) Delete(ns, file string) error {
	log := f.log.Named("Delete")

	p := path.Join(f.root, ns, file)

	err := os.Remove(p)
	if err != nil {
		log.Error("failed to delete file", zap.Error(err))
		return err
	}
	return nil
}

func (f *FSHandler) DeleteNS(ns string) error {
	log := f.log.Named("DeleteNS")

	p := path.Join(f.root, ns)

	err := os.RemoveAll(p)
	if err != nil {
		log.Error("failed to delete namespace", zap.Error(err))
		return err
	}
	return nil
}

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
package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/infinimesh/http-fs/pkg/io/fs"
	"github.com/infinimesh/http-fs/pkg/router"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
    log *zap.Logger
)

func init() {
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	log = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
		
	flag.Int("log-level", 0, "Log level between -1 and 5, where -1 is Debug and 5 is Fatal")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

    atom.SetLevel(
		zapcore.Level(viper.GetInt("log-level")),
	)
}

func main() {
    defer log.Sync()
    log.Info("Starting infinimesh HTTP FileServer")

    handler := fs.NewFileSystemHandler(log, "static")
    mux := router.NewRouter(handler)

    srv := &http.Server{
        Handler: mux,
        Addr:    ":8000",
    }

    log.Info("Starting HTTP Server")
    log.Fatal("Failed to serve", zap.Error(srv.ListenAndServe()))
}
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
package mw

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/infinimesh/proto/node"
	accpb "github.com/infinimesh/proto/node/access"
	nspb "github.com/infinimesh/proto/node/namespaces"
)

func InfinimeshMiddleware(logger *zap.Logger, host string) func(next http.Handler) http.Handler {

	log := logger.Named("infinimesh")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx, host, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("Could not connect to API", zap.Error(err))
	}

	api := pb.NewNamespacesServiceClient(conn)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			vars := mux.Vars(r)
			log.Debug("Request", zap.Any("vars", vars), zap.Bool("unauthorized", bearer == ""))

			ctx := r.Context()

			// Setting default access level to be rw
			acc := Access{
				Read:  true,
				Write: false,
			}

			// Restricting read access to namespace level methods as well in case request is unauthorized
			// This is to make files public by default but prevent seeing files overview
			_, ok := vars["file"]
			if !ok {
				acc.Read = false
			}

			if bearer == "" {
				// Skipping namespaces check if no token present
				ctx = context.WithValue(ctx, AccessKey, acc)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			uuid, ok := vars["ns"]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Requesting namespace information from infinimesh Repo
			ctx = metadata.AppendToOutgoingContext(context.Background(), "authorization", bearer)

			ns, err := api.Get(ctx, &nspb.Namespace{Uuid: uuid})
			if err != nil {
				log.Error("could not get namespace", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if ns.Access == nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if ns.Access.Level < accpb.Level_READ {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			acc.Read = true
			if ns.Access.Level >= accpb.Level_ADMIN {
				acc.Write = true
			}

			ctx = context.WithValue(ctx, AccessKey, acc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

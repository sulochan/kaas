package api

import (
	"net/http"

	"github.com/gorilla/context"

	"github.com/sulochan/kaas/models"
)

func SetContext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		opts := models.AuthOpts{}
		opts.Type = r.Header.Get("X-Auth-Type")
		opts.Token = r.Header.Get("X-Auth-Token")
		opts.Password = r.Header.Get("X-Auth-Password")
		opts.Username = r.Header.Get("X-Auth-Username")
		opts.ProjectId = r.Header.Get("X-Auth-ProjectId")

		context.Set(r, "authOpts", opts)
		context.Set(r, "projectid", opts.ProjectId)
		context.Set(r, "username", opts.Username)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

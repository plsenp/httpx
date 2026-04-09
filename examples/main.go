package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/plsenp/httpx"
	"github.com/plsenp/httpx/openapi"
)

func main() {
	srv := httpx.NewServer(httpx.WithOpenAPISpec(openapi.Config{
		Title:       "Auth API",
		Version:     "1.0.0",
		Description: "Auth API",
	}))
	r := srv.Route()
	// b := &Biz{
	// 	data: make(map[string]string),
	// }

	r.POST("/test/{id}", func(ctx *httpx.Ctx) error {
		type TestReq struct {
			ID   string `path:"id"`
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		var in TestReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		w := ctx.Writer()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		ctx.JSON(300, in) //nolint:errcheck
		return nil
	}, logger())
	srv.Start() //nolint:errcheck
}

type Biz struct {
	data map[string]string
}

type Auth interface {
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
	Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error)
}

func logger() httpx.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("before", "header", w.Header())
			h.ServeHTTP(w, r)
			slog.Info("after", "header", w.Header())
		})
	}
}

type (
	LoginReq struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	LoginResp struct {
		Token string `json:"token"`
	}

	RegisterReq struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
		Confirm  string `json:"confirm" validate:"required,eqfield=Password"`
	}

	RegisterResp struct {
		Token string `json:"token"`
	}
)

func (b *Biz) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	if (req.Username != "admin" || req.Password != "123456") && b.data[req.Username] != req.Password {
		return nil, errors.New("invalid username or password")
	}
	return &LoginResp{
		Token: "123456",
	}, nil
}

func (b *Biz) Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error) {
	if _, ok := b.data[req.Username]; ok {
		return nil, errors.New("username already exists")
	}
	b.data[req.Username] = req.Password
	return &RegisterResp{
		Token: "123456",
	}, nil
}

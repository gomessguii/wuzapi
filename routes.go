package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"
	"wuzapi/controllers/chat"
	"wuzapi/controllers/group"
	"wuzapi/controllers/session"
	"wuzapi/controllers/user"
	"wuzapi/controllers/webhook"
	"wuzapi/internal/controller"
	internalTypes "wuzapi/internal/types"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type Middleware = alice.Constructor

func setupRoutes(s *controller.Server) {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	if *logType == "json" {
		log = zerolog.New(os.Stdout).With().Timestamp().Str("role", filepath.Base(os.Args[0])).Str("host", *address).Logger()
	} else {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		log = zerolog.New(output).With().Timestamp().Str("role", filepath.Base(os.Args[0])).Str("host", *address).Logger()
	}

	c := alice.New()
	c = c.Append(s.Authalice)
	c = c.Append(hlog.NewHandler(log))

	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Str("userid", r.Context().Value("userinfo").(internalTypes.Values).Get("Id")).
			Msg("Got API Request")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	sessionController := &session.SessionController{Server: s}
	sessionController.SignRoutes(c)

	webhookController := &webhook.WebhookController{Server: s}
	webhookController.SignRoutes(c)

	userController := &user.UserController{Server: s}
	userController.SignRoutes(c)

	chatController := &chat.ChatController{Server: s}
	chatController.SignRoutes(c)

	chatMessageController := &chat.ChatMessageController{Server: s}
	chatMessageController.SignRoutes(c)

	groupController := &group.GroupController{Server: s}
	groupController.SignRoutes(c)

	s.Router.PathPrefix("/").Handler(http.FileServer(http.Dir(exPath + "/static/")))
}

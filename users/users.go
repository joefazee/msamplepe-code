package users

import (
	"github.com/timchuks/monieverse/core/server"
)

type usersController struct {
	srv *server.Server
}

func NewUsersController(srv *server.Server) *usersController {
	return &usersController{
		srv: srv,
	}
}

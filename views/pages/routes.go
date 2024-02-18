package pages

import (
	"net/http"
	"townwatch/views/pages/join"

	"github.com/a-h/templ"
)

func RegisterPagesRoutes() {

	http.Handle("/join", templ.Handler(join.JoinPage()))

}

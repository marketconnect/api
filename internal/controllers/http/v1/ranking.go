package v1

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	ranking = "/ranking"
)

type rankingHandler struct {
}

func NewRankingHandler() *rankingHandler {
	return &rankingHandler{}
}

func (h *rankingHandler) Register(router *httprouter.Router) {
	router.GET(ranking, jwtMiddleware(handlerFunc(h.GetAllRanking)))

}

func (h *rankingHandler) GetAllRanking(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Context().Value("username") == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	currentUser := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Welcome, %s!", currentUser)
}

package rest

import (
	"net/http"

	"bitbucket.org/kleinnic74/photos/swarm"
	"github.com/gorilla/mux"
)

type PeersAPI struct {
	peers *swarm.Controller
}

func NewPeersAPI(peers *swarm.Controller) *PeersAPI {
	return &PeersAPI{peers: peers}
}

func (p *PeersAPI) InitRoutes(router *mux.Router) {
	router.HandleFunc("/peers", p.listPeers).Methods(http.MethodGet)
}

func (p *PeersAPI) listPeers(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Services  []string     `json:"services,omitempty"`
		Instances []swarm.Peer `json:"instances,omitempty"`
	}{
		Services:  p.peers.GetSeenServices(),
		Instances: p.peers.GetPeers(),
	}
	Respond(r).WithJSON(w, http.StatusOK, &simplePayload{Data: data})
}

package adapter

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"telegram/internal/domain"
	"telegram/internal/service"
)

type RestHandler struct {
	service *service.Service
}

func NewRestHandler(service *service.Service) *RestHandler {
	return &RestHandler{service: service}
}

func (r *RestHandler) Handle(w http.ResponseWriter, req *http.Request) {
	var update domain.Update

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("could not read request body %s", err.Error())
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Printf("request body is %s", string(body))

	// if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
	// 	log.Printf("error decoding message %s", err.Error())
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("error unmarshalling message %s", err.Error())
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := r.service.ReceiveMessage(update.Message); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

package Response

import (
	"encoding/json"
	"net/http"

	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

type Server = server.Server

type ServerResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Id      int       `json:"id,omitempty"`
	Data    []*Server `json:"data,omitempty"`
}

func WriteSuccessResponse(w http.ResponseWriter, status int, server *Server) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := ServerResponse{
		Status:  "success",
		Message: "Server added successfully",
		Id:      server.Id,
	}
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, encodeErr.Error())
	}
}

func WriteSuccessResponseArray(w http.ResponseWriter, status int, servers []*Server) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := ServerResponse{
		Status:  "success",
		Message: "Server added successfully",
		Data:    servers,
	}
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, encodeErr.Error())
	}
}

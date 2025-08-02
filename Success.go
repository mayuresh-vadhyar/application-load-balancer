package main

import (
	"encoding/json"
	"net/http"
)

type ServerResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Id      int       `json:"id"`
	Data    []*Server `json:"data"`
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

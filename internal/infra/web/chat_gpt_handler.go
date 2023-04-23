package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/robertomorel/chatservice/internal/usecase/chatcompletion"
)

type WebChatGPTHandler struct {
	CompletionUseCase chatcompletion.ChatCompletionUseCase //
	Config            chatcompletion.ChatCompletionConfigInputDTO
	AuthToken         string
}

// Constructor
func NewWebChatGPTHandler(usecase chatcompletion.ChatCompletionUseCase, config chatcompletion.ChatCompletionConfigInputDTO, authToken string) *WebChatGPTHandler {
	return &WebChatGPTHandler{
		CompletionUseCase: usecase,
		Config:            config,
		AuthToken:         authToken,
	}
}

// Padrão do webserver do Go
func (h *WebChatGPTHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Post é o único método permitido
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Checando se o token é válido
	if r.Header.Get("Authorization") != h.AuthToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Tomando todas as informações do body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validando se body é um json
	if !json.Valid(body) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var dto chatcompletion.ChatCompletionInputDTO
	// O "Unmarshal" está pegando todos os dados do body e jogando para o &dto
	err = json.Unmarshal(body, &dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dto.Config = h.Config

	result, err := h.CompletionUseCase.Execute(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	tiktoken_go "github.com/j178/tiktoken-go"
)

type Message struct {
	ID        string    // Identificador
	Role      string    // User | Assistant | System
	Content   string    // Conteúdo
	Tokens    int       // Qnde de tokens da mensagem
	Model     *Model    // Baseado no modelo, podemos fazer a contagem dos tokens
	CreatedAt time.Time // Timestamp
}

// * representa um ponteiro para a memória
// & representa um endereço de memória
// É preciso validar a quantidade de tokens para que não exceda o limite do modelo (ChatGPT 3.5 turbo tem 4096 tokens, por exemplo)
func NewMessage(role, content string, model *Model) (*Message, error) {
	// Pacote para contar a quantidade de tokens a partir do model
	totalTokens := tiktoken_go.CountTokens(model.GetModelName(), content)

	msg := &Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Tokens:    totalTokens,
		Model:     model,
		CreatedAt: time.Now(),
	}

	// Validando a mensagem antes de retornar
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *Message) Validate() error {
	// Regra 1: Role
	if m.Role != "user" && m.Role != "system" && m.Role != "assistant" {
		return errors.New("invalid role")
	}

	// Regra 2: Não pode conteúdo em branco
	if m.Content == "" {
		return errors.New("content is empty")
	}

	// Regra 3: Created At em branco
	if m.CreatedAt.IsZero() {
		return errors.New("invalid created at")
	}
	return nil
}

// Qde de tokens da mensagem
func (m *Message) GetQtdTokens() int {
	return m.Tokens
}

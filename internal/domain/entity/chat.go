package entity

import (
	"errors"

	"github.com/google/uuid"
)

// Configurações copiadas do site do OpenAI
type ChatConfig struct {
	Model            *Model
	Temperature      float32  // 0.0 to 1.0 - Precisão da resposta
	TopP             float32  // 0.0 to 1.0 - to a low value, like 0.1, the model will be very conservative in its word choices, and will tend to generate relatively predictable prompts
	N                int      // number of messages to generate
	Stop             []string // list of tokens to stop on
	MaxTokens        int      // number of tokens to generate
	PresencePenalty  float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics.
	FrequencyPenalty float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, increasing the model's likelihood to talk about new topics.
}

type Chat struct {
	ID                   string      // Identificador
	UserID               string      // User
	InitialSystemMessage *Message    // Mensagem inicial do system
	Messages             []*Message  // List de mensagens
	ErasedMessages       []*Message  // Apagando mensagens antigas quando necessário
	Status               string      // Saber se o chat foi finalizado ou não
	TokenUsage           int         // Qntos tokens foram usados
	Config               *ChatConfig // Configuração do chat
}

func NewChat(userID string, initialSystemMessage *Message, chatConfig *ChatConfig) (*Chat, error) {
	chat := &Chat{
		ID:                   uuid.New().String(),
		UserID:               userID,
		InitialSystemMessage: initialSystemMessage,
		Status:               "active",
		Config:               chatConfig,
		TokenUsage:           0,
	}
	// Adiciona a mensagem inicial
	chat.AddMessage(initialSystemMessage)

	if err := chat.Validate(); err != nil {
		return nil, err
	}
	return chat, nil
}

func (c *Chat) Validate() error {
	if c.UserID == "" {
		return errors.New("user id is empty")
	}
	if c.Status != "active" && c.Status != "ended" {
		return errors.New("invalid status")
	}
	if c.Config.Temperature < 0 || c.Config.Temperature > 2 {
		return errors.New("invalid temperature")
	}
	// ... more validations for config
	return nil
}

func (c *Chat) AddMessage(m *Message) error {
	// Valida se mensagem já está terminada
	if c.Status == "ended" {
		return errors.New("chat is ended. no more messages allowed")
	}

	for {
		// Se a qnde máxima permitida de tokens do modelo for maior que a quantidade de tokens da mensagem + tokens usados no chat...
		if c.Config.Model.GetMaxTokens() >= m.GetQtdTokens()+c.TokenUsage {
			// Adiciona a nova mensagem no chat
			c.Messages = append(c.Messages, m)
			// Atualiza a qnde de tokens
			c.RefreshTokenUsage()
			break
		}
		// Else...
		// Pegando a mensagem mais antiga e adiciona na lista ErasedMessages
		c.ErasedMessages = append(c.ErasedMessages, c.Messages[0])
		// Apaga a última mensagem
		c.Messages = c.Messages[1:]
		c.RefreshTokenUsage()
	}
	return nil
}

func (c *Chat) GetMessages() []*Message {
	return c.Messages
}

func (c *Chat) CountMessages() int {
	return len(c.Messages)
}

func (c *Chat) End() {
	c.Status = "ended"
}

// Atualiza a quantidade de tokens
func (c *Chat) RefreshTokenUsage() {
	c.TokenUsage = 0
	// Laço por todas as mensagens do chat
	for m := range c.Messages {
		c.TokenUsage += c.Messages[m].GetQtdTokens()
	}
}

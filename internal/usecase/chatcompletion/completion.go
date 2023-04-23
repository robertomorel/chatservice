/*
	Toda vez que alguém tiver mandando mensagem no chat, a IA vai completando as mensagem e continuando
*/

package chatcompletion

import (
	"context"
	"errors"

	"github.com/robertomorel/chatservice/internal/domain/entity"
	"github.com/robertomorel/chatservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)

// Dados de completion para configuração
type ChatCompletionConfigInputDTO struct {
	Model                string
	ModelMaxTokens       int
	Temperature          float32  // 0.0 to 1.0
	TopP                 float32  // 0.0 to 1.0 - to a low value, like 0.1, the model will be very conservative in its word choices, and will tend to generate relatively predictable prompts
	N                    int      // number of messages to generate
	Stop                 []string // list of tokens to stop on
	MaxTokens            int      // number of tokens to generate
	PresencePenalty      float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics.
	FrequencyPenalty     float32  // -2.0 to 2.0 - Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, increasing the model's likelihood to talk about new topics.
	InitialSystemMessage string
}

// Type para dados de entrada
type ChatCompletionInputDTO struct {
	ChatID      string                       `json:"chat_id,omitempty"`
	UserID      string                       `json:"user_id"`
	UserMessage string                       `json:"user_message"`
	Config      ChatCompletionConfigInputDTO `json:"config"`
}

type ChatCompletionOutputDTO struct {
	ChatID  string `json:"chat_id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"` // Resposta do ChatGPT
}

type ChatCompletionUseCase struct {
	ChatGateway  gateway.ChatGateway // Trabalhar com inversão de controle. Salvar dados no DB.
	OpenAIClient *openai.Client      // Client da OpenAI. Chamar a API para fazer a completion
	// Stream       chan ChatCompletionOutputDTO // Canal de comunicação do tipo ChatCompletionOutputDTO
}

// Criando um novo CompletionUseCase
// func NewChatCompletionUseCase(chatGateway gateway.ChatGateway, openAIClient *openai.Client, stream chan ChatCompletionOutputDTO) *ChatCompletionUseCase {
func NewChatCompletionUseCase(chatGateway gateway.ChatGateway, openAIClient *openai.Client) *ChatCompletionUseCase {
	return &ChatCompletionUseCase{
		ChatGateway:  chatGateway,
		OpenAIClient: openAIClient,
		// Stream:       stream,
	}
}

/*
[ctx context.Context]
O contexto funciona para passar dados, header entre locais e também possui recursos de cancelamento
*/
func (uc *ChatCompletionUseCase) Execute(ctx context.Context, input ChatCompletionInputDTO) (*ChatCompletionOutputDTO, error) {
	// Verifica se o chat já existe
	chat, err := uc.ChatGateway.FindChatByID(ctx, input.ChatID)
	if err != nil {
		// Se chat não existir
		if err.Error() == "chat not found" {
			// Cria novo chat entity
			chat, err = createNewChat(input)
			// Se deu erro ao criar o chat...
			if err != nil {
				return nil, errors.New("error creating new chat: " + err.Error())
			}
			// Persiste o chat no DB. Usando o Gateway.
			err = uc.ChatGateway.CreateChat(ctx, chat)
			// Erro ao salvar o chat no DB
			if err != nil {
				return nil, errors.New("error persisting new chat: " + err.Error())
			}
		} else {
			// Erro na biblioteca ao buscar o chat, não na aplicação
			return nil, errors.New("error fetching existing chat: " + err.Error())
		}
	}

	// Criando nova mensagem de usuário
	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating new message: " + err.Error())
	}

	// Adiciona a nova mensagem ao chat
	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("error adding new message: " + err.Error())
	}

	// Instanciando variável do tipo array de openai.ChatCompletionMessage{} para guardar todas as mensagens do chat
	messages := []openai.ChatCompletionMessage{}
	// Percorrendo todas as mensagens do chat
	for _, msg := range chat.Messages {
		// Criando arr de mensagens no formato reconhecido pela API do OpenAI
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Chamando a API para fazer o completion
	resp, err := uc.OpenAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.Name,
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			PresencePenalty:  chat.Config.PresencePenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
			Stop:             chat.Config.Stop,
		},
	)
	if err != nil {
		return nil, errors.New("error openai: " + err.Error())
	}

	assistant, err := entity.NewMessage("assistant", resp.Choices[0].Message.Content, chat.Config.Model)
	if err != nil {
		return nil, err
	}
	err = chat.AddMessage(assistant)
	if err != nil {
		return nil, err
	}

	err = uc.ChatGateway.SaveChat(ctx, chat)
	if err != nil {
		return nil, err
	}

	output := &ChatCompletionOutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: resp.Choices[0].Message.Content,
	}

	return output, nil

	// STREAM -------------------------------------------------------/
	// resp, err := uc.OpenAIClient.CreateChatCompletionStream(
	// 	context.Background(),
	// 	openai.ChatCompletionRequest{
	// 		Model:            chat.Config.Model.Name,
	// 		Messages:         messages,
	// 		MaxTokens:        chat.Config.MaxTokens,
	// 		Temperature:      chat.Config.Temperature,
	// 		TopP:             chat.Config.TopP,
	// 		PresencePenalty:  chat.Config.PresencePenalty,
	// 		FrequencyPenalty: chat.Config.FrequencyPenalty,
	// 		Stop:             chat.Config.Stop,
	// 		Stream:           true, // Para receber respostas conforme for enviando mensagens
	// 	},
	// )
	// if err != nil {
	// 	return nil, errors.New("error openai: " + err.Error())
	// }

	// var fullResponse strings.Builder

	// for {
	// 	response, err := resp.Recv() // Recebendo dados por streaming
	// 	// Se a mensagem acabou...
	// 	if errors.Is(err, io.EOF) {
	// 		break
	// 	}
	// 	if err != nil {
	// 		return nil, errors.New("error streaming response: " + err.Error())
	// 	}

	// 	fullResponse.WriteString(response.Choices[0].Delta.Content)
	// 	r := &ChatCompletionOutputDTO{
	// 		ChatID:  chat.ID,
	// 		UserID:  input.UserID,
	// 		Content: fullResponse.String(),
	// 	}

	// 	// Mandando a resposta pro canal de stream e no GRPC pegamos os dados
	// 	// Mandar informações de uma thread para outra
	// 	uc.Stream <- *r
	// }

	// // Guardar a mensagem no DB
	// assistant, err := entity.NewMessage("assistant", fullResponse.String(), chat.Config.Model)
	// if err != nil {
	// 	return nil, err
	// }
	// // Resposta do chatGPT sendo add
	// err = chat.AddMessage(assistant)
	// if err != nil {
	// 	return nil, err
	// }

	// // Salvando o chat
	// err = uc.ChatGateway.SaveChat(ctx, chat)
	// if err != nil {
	// 	return nil, err
	// }

	// output := &ChatCompletionOutputDTO{
	// 	ChatID:  chat.ID,
	// 	UserID:  input.UserID,
	// 	Content: fullResponse.String(),
	// }

	// return output, nil
	// STREAM -------------------------------------------------------/
}

// Criando uma nova entidade de chat
func createNewChat(input ChatCompletionInputDTO) (*entity.Chat, error) {
	model := entity.NewModel(input.Config.Model, input.Config.ModelMaxTokens)
	chatConfig := &entity.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
		Model:            model,
	}

	initialMessage, err := entity.NewMessage("system", input.Config.InitialSystemMessage, model)
	if err != nil {
		return nil, errors.New("error creating initial message: " + err.Error())
	}
	chat, err := entity.NewChat(input.UserID, initialMessage, chatConfig)
	if err != nil {
		return nil, errors.New("error creating new chat: " + err.Error())
	}
	return chat, nil
}

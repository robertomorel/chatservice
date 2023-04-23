package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robertomorel/chatservice/configs"
	"github.com/robertomorel/chatservice/internal/infra/grpc/server"
	"github.com/robertomorel/chatservice/internal/infra/repository"
	"github.com/robertomorel/chatservice/internal/infra/web"
	"github.com/robertomorel/chatservice/internal/infra/web/webserver"
	"github.com/robertomorel/chatservice/internal/usecase/chatcompletion"
	"github.com/robertomorel/chatservice/internal/usecase/chatcompletionstream"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Preparando variáveis de configuração
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	// Abrindo configuração
	conn, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	// Depois que roda todo o main, o "defer" fecha a conexão
	defer conn.Close()

	// Instanciando o repositório
	repo := repository.NewChatRepositoryMySQL(conn)
	// Instanciando o client do OpenAI
	client := openai.NewClient(configs.OpenAIApiKey)

	// Criando as configurações do OpenAI
	chatConfig := chatcompletion.ChatCompletionConfigInputDTO{
		Model:                configs.Model,
		ModelMaxTokens:       configs.ModelMaxTokens,
		Temperature:          float32(configs.Temperature),
		TopP:                 float32(configs.TopP),
		N:                    configs.N,
		Stop:                 configs.Stop,
		MaxTokens:            configs.MaxTokens,
		InitialSystemMessage: configs.InitialChatMessage,
	}

	// Criando as configurações do OpenAI para stream
	chatConfigStream := chatcompletionstream.ChatCompletionConfigInputDTO{
		Model:                configs.Model,
		ModelMaxTokens:       configs.ModelMaxTokens,
		Temperature:          float32(configs.Temperature),
		TopP:                 float32(configs.TopP),
		N:                    configs.N,
		Stop:                 configs.Stop,
		MaxTokens:            configs.MaxTokens,
		InitialSystemMessage: configs.InitialChatMessage,
	}

	usecase := chatcompletion.NewChatCompletionUseCase(repo, client)

	// Criando stream channel
	streamChannel := make(chan chatcompletionstream.ChatCompletionOutputDTO)
	usecaseStream := chatcompletionstream.NewChatCompletionUseCase(repo, client, streamChannel)

	fmt.Println("Starting gRPC server on port " + configs.GRPCServerPort)
	// Instanciando o servidor gRPC
	grpcServer := server.NewGRPCServer(
		*usecaseStream,
		chatConfigStream,
		configs.GRPCServerPort,
		configs.AuthToken,
		streamChannel,
	)
	// Criando uma nova thread com o servidor do grpc
	go grpcServer.Start()

	// Criando o webserver
	webserver := webserver.NewWebServer(":" + configs.WebServerPort)
	webserverChatHandler := web.NewWebChatGPTHandler(*usecase, chatConfig, configs.AuthToken)
	// Adicionando o primeiro handler, com o path /chat
	webserver.AddHandler("/chat", webserverChatHandler.Handle)

	fmt.Println("Server running on port " + configs.WebServerPort)
	webserver.Start()
}

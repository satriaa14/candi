package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	_ "agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/Shopify/sarama"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// App service
type App struct {
	service       factory.ServiceFactory
	httpServer    *echo.Echo
	grpcServer    *grpc.Server
	kafkaConsumer sarama.ConsumerGroup
}

// New service app
func New(service factory.ServiceFactory) *App {
	defer log.Printf("Starting %s service\n", service.Name())

	// load json schema for document validation
	// jsonschema.Load(string(service.Name()))

	dependency := service.GetDependency()

	appInstance := new(App)
	appInstance.service = service

	if config.BaseEnv().UseHTTP {
		appInstance.httpServer = echo.New()
	}

	if config.BaseEnv().UseGRPC {
		// init grpc server
		appInstance.grpcServer = grpc.NewServer(
			grpc.MaxSendMsgSize(200*int(helper.MByte)), grpc.MaxRecvMsgSize(200*int(helper.MByte)),
			grpc.UnaryInterceptor(dependency.Middleware.GRPCAuth),
			grpc.StreamInterceptor(dependency.Middleware.GRPCAuthStream),
		)
	}

	if config.BaseEnv().UseGraphQL {
		gqlHandler := appInstance.graphqlHandler(dependency.Middleware)
		appInstance.httpServer.Add(http.MethodGet, "/graphql", echo.WrapHandler(gqlHandler))
		appInstance.httpServer.Add(http.MethodPost, "/graphql", echo.WrapHandler(gqlHandler))
		appInstance.httpServer.GET("/graphql/playground", gqlHandler.servePlayground)
	}

	if config.BaseEnv().UseKafka {
		// init kafka consumer
		kafkaConsumer, err := sarama.NewConsumerGroup(config.BaseEnv().Kafka.Brokers,
			config.BaseEnv().Kafka.ConsumerGroup,
			dependency.Config.KafkaConsumerConfig)
		if err != nil {
			log.Panicf("Error creating consumer group client: %v", err)
		}
		appInstance.kafkaConsumer = kafkaConsumer
	}

	return appInstance
}

// Run start app
func (a *App) Run(ctx context.Context) {

	hasServiceHandlerRunning := a.httpServer != nil || a.grpcServer != nil || a.kafkaConsumer != nil
	if !hasServiceHandlerRunning {
		panic("No service handler running")
	}

	// serve http server
	go a.ServeHTTP()

	// serve grpc server
	go a.ServeGRPC()

	// serve kafka consumer
	go a.KafkaConsumer()

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal

	a.Shutdown(ctx)
}

// Shutdown graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) Shutdown(ctx context.Context) {
	println()

	if a.httpServer != nil {
		log.Println("Stopping HTTP server...")
		if err := a.httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
	}

	if a.grpcServer != nil {
		log.Println("Stopping GRPC server...")
		a.grpcServer.GracefulStop()
	}

	if a.kafkaConsumer != nil {
		log.Println("Stopping kafka consumer...")
		a.kafkaConsumer.Close()
	}
}

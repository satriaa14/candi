package invitation

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/invitation/delivery"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

const (
	// Invitation service name
	Invitation constant.Module = "Invitation"
)

// Module model
type Module struct {
	restHandler    *delivery.RestInvitationHandler
	graphqlHandler *delivery.GraphQLHandler
	kafkaHandler   *delivery.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestInvitationHandler(deps.Middleware)
	mod.graphqlHandler = delivery.NewGraphQLHandler(deps.Middleware)
	mod.kafkaHandler = delivery.NewKafkaHandler([]string{"test", "coba"})
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = m.restHandler
	case helper.V2:
		d = nil // TODO versioning
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Invitation), m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberHandler {
	switch subsType {
	case constant.Kafka:
		return m.kafkaHandler
	}
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Invitation
}

package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/Shopify/sarama"
)

// KafkaConsumer consume data from kafka
func (a *App) KafkaConsumer() {
	if a.kafkaConsumer == nil {
		return
	}

	topicInfo := make(map[string][]string)
	var handlers = make(map[string][]interfaces.SubscriberHandler)
	for _, m := range a.service.GetModules() {
		if h := m.SubscriberHandler(constant.Kafka); h != nil {
			for _, topic := range h.GetTopics() {
				handlers[topic] = append(handlers[topic], h) // one same topic consumed by multiple module
				topicInfo[topic] = append(topicInfo[topic], string(m.Name()))
			}
		}
	}
	consumer := kafkaConsumer{
		handlers: handlers,
	}

	println(helper.StringYellow("Kafka consumer is active"))
	var consumeTopics []string
	for topic, handlerNames := range topicInfo {
		print(helper.StringYellow(fmt.Sprintf("[KAFKA-CONSUMER] (topic): %-8s --> (modules): [%s]\n", topic, strings.Join(handlerNames, ", "))))
		consumeTopics = append(consumeTopics, topic)
	}
	println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.kafkaConsumer.Consume(ctx, consumeTopics, &consumer); err != nil {
		log.Printf("Error from consumer: %v", err)
	}
}

// kafkaConsumer represents a Sarama consumer group consumer
type kafkaConsumer struct {
	handlers map[string][]interfaces.SubscriberHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)

		for _, handler := range c.handlers[message.Topic] {
			handler.ProcessMessage(session.Context(), message.Value)
		}

		session.MarkMessage(message, "")
	}

	return nil
}

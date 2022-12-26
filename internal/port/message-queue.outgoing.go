package port

import "telegram/internal/domain"

type OutgoingMessageQueue interface {
	PublishMessage(message domain.Message) error
}

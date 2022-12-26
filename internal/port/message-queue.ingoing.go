package port

import "telegram/internal/domain"

type IngoingMessageQueue interface {
	StreamMessages() (<-chan domain.Message, error)
}

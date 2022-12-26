package port

import "telegram/internal/domain"

type Repository interface {
	SaveMessage(message domain.Message) error
}

package port

type MemoryRepository interface {
	SaveLatestSentMessageTimestamp(userId int, timestamp int64) error
	GetLatestSentMessageTimestamp(userId int) (int64, error)
}

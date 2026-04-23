package output

import "context"

// EventPublisher define a interface para publicação de eventos de domínio.
type EventPublisher interface {
	// PublishUserUpdated publica um evento user.updated no Redis Pub/Sub.
	PublishUserUpdated(ctx context.Context, userID string, updatedFields map[string]any) error
}

package repo

import "context"

type NotificationService interface {
    Send(ctx context.Context, userID int64, title, body string) error
}
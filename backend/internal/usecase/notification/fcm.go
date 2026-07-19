package usecase

import (
	"context"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type TokenRepository interface {
	GetTokensByUserID(ctx context.Context, userID int64) ([]string, error)
	DeleteToken(ctx context.Context, token string) error
}

type FCMNotificationService struct {
	fcmClient *messaging.Client
	tokenRepo TokenRepository
}

func NewFCMNotificationService(ctx context.Context, credentialsPath string, tokenRepo TokenRepository) (*FCMNotificationService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMNotificationService{
		fcmClient: client,
		tokenRepo: tokenRepo,
	}, nil
}

func (s *FCMNotificationService) Send(ctx context.Context, userID int64, title, body string) error {
	tokens, err := s.tokenRepo.GetTokensByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	br, err := s.fcmClient.SendEachForMulticast(ctx, message)
	if err != nil {
		return err
	}

	if br.FailureCount > 0 {
		for idx, resp := range br.Responses {
			if !resp.Success {
				slog.Warn("failed to send push to token", "token", tokens[idx], "error", resp.Error)
				s.tokenRepo.DeleteToken(ctx, tokens[idx])
			}
		}
	}

	return nil
}
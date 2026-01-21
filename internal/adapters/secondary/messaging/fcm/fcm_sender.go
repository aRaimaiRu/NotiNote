package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// FCMSender implements the NotificationSender interface using Firebase Cloud Messaging
type FCMSender struct {
	client *messaging.Client
	logger *logrus.Logger
}

// NewFCMSender creates a new FCM sender
func NewFCMSender(credentialsFile string, logger *logrus.Logger) (*FCMSender, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	logger.Info("FCM sender initialized successfully")

	return &FCMSender{
		client: client,
		logger: logger,
	}, nil
}

// SendPushNotification sends a push notification to a single device
func (s *FCMSender) SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		// Web push configuration
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Icon:  "/icons/notification-icon.png",
			},
			FCMOptions: &messaging.WebpushFCMOptions{
				Link: data["click_url"],
			},
		},
		// Android configuration
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title:       title,
				Body:        body,
				Sound:       "default",
				ChannelID:   "note_reminders",
				ClickAction: "OPEN_NOTE",
			},
		},
		// iOS configuration
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
					Badge: func() *int { i := 1; return &i }(),
				},
			},
		},
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"device_token": deviceToken[:min(20, len(deviceToken))] + "...",
			"title":        title,
		}).Error("Failed to send FCM message")
		return fmt.Errorf("failed to send FCM message: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"message_id": response,
		"title":      title,
	}).Info("FCM message sent successfully")

	return nil
}

// SendToMultipleDevices sends a push notification to multiple devices
func (s *FCMSender) SendToMultipleDevices(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) error {
	if len(deviceTokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: deviceTokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		// Web push configuration
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Icon:  "/icons/notification-icon.png",
			},
		},
		// Android configuration
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title:     title,
				Body:      body,
				Sound:     "default",
				ChannelID: "note_reminders",
			},
		},
		// iOS configuration
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
				},
			},
		},
	}

	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"device_count": len(deviceTokens),
			"title":        title,
		}).Error("Failed to send multicast FCM message")
		return fmt.Errorf("failed to send multicast FCM message: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"success_count": response.SuccessCount,
		"failure_count": response.FailureCount,
		"title":         title,
	}).Info("Multicast FCM message sent")

	// Log individual failures
	if response.FailureCount > 0 {
		for i, sendResponse := range response.Responses {
			if sendResponse.Error != nil {
				s.logger.WithError(sendResponse.Error).WithFields(logrus.Fields{
					"token_index": i,
				}).Warn("Individual FCM send failed")
			}
		}
	}

	return nil
}

// BatchResponse represents the result of a batch send operation
type BatchResponse struct {
	SuccessCount int
	FailureCount int
	Responses    []*messaging.SendResponse
}

// SendBatchWithResponse sends to multiple devices and returns detailed response
func (s *FCMSender) SendBatchWithResponse(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) (*BatchResponse, error) {
	if len(deviceTokens) == 0 {
		return &BatchResponse{}, nil
	}

	message := &messaging.MulticastMessage{
		Tokens: deviceTokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send multicast FCM message: %w", err)
	}

	return &BatchResponse{
		SuccessCount: response.SuccessCount,
		FailureCount: response.FailureCount,
		Responses:    response.Responses,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

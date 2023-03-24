package bot

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type ImageService struct {
	client *openai.Client
}

func NewImageService(key string) *ImageService {
	return &ImageService{
		client: openai.NewClient(key),
	}
}

func (s *ImageService) HandleImageCommand(c tele.Context) error {
	prompt := strings.TrimPrefix(c.Message().Text, "/image ")

	request := openai.ImageRequest{
		Prompt:         prompt,
		N:              1,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
	}

	log.Infof("creating image with prompt: %s", prompt)
	resp, err := s.client.CreateImage(context.Background(), request)
	if err != nil {
		return err
	}

	log.Infof("sending image: %s", resp.Data[0].URL)
	photo := tele.Photo{File: tele.FromURL(resp.Data[0].URL)}
	_, err = photo.Send(c.Bot(), c.Message().Sender, &tele.SendOptions{
		ReplyTo: c.Message(),
	})

	return err
}

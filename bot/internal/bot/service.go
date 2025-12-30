package bot

import (
	"fmt"
	"image"
	"log"

	tele "gopkg.in/telebot.v3"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/config"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/face"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
)

// Service представляет сервис бота.
type Service struct {
	bot      *tele.Bot
	config   *config.Config
	detector face.Detector
	marker   func(img image.Image, faces []face.Face) (image.Image, error)
	voting   *voting.Service
}

// NewService создает новый сервис бота.
func NewService(
	cfg *config.Config,
	detector face.Detector,
	votingService *voting.Service,
) (*Service, error) {
	pref := tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	service := &Service{
		bot:      bot,
		config:   cfg,
		detector: detector,
		marker:   face.MarkFaces,
		voting:   votingService,
	}

	return service, nil
}

// Start запускает бота.
func (s *Service) Start() {
	s.registerHandlers()
	log.Println("Bot started")
	s.bot.Start()
}

// Stop останавливает бота.
func (s *Service) Stop() {
	s.bot.Stop()
}

// SetVotingService устанавливает сервис голосований.
func (s *Service) SetVotingService(votingService *voting.Service) {
	s.voting = votingService
}

// PublishResults публикует результаты голосования (для использования в callback).
func (s *Service) PublishResults(session *voting.Session) error {
	return s.publishResults(session)
}

// registerHandlers регистрирует обработчики событий.
func (s *Service) registerHandlers() {
	s.bot.Handle(tele.OnPhoto, s.handlePhoto)
	s.bot.Handle(tele.OnCallback, s.handleCallback)
}


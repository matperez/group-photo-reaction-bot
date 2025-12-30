package bot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
	imgproc "github.com/matperez/group-photo-reaction-bot/bot/pkg/image"
)

// handlePhoto обрабатывает получение фото в чате.
func (s *Service) handlePhoto(c tele.Context) error {
	// Проверяем, что это групповой чат
	if c.Chat().Type != tele.ChatGroup && c.Chat().Type != tele.ChatSuperGroup {
		return nil
	}

	photo := c.Message().Photo
	if photo == nil {
		return nil
	}

	// Проверяем rate limiting
	if !s.voting.Store().CanCreateSession(c.Chat().ID, 1*time.Hour) {
		return nil
	}

	// Загружаем фото
	file, err := s.bot.File(&photo.File)
	if err != nil {
		log.Printf("Failed to download photo: %v", err)
		return nil
	}

	photoData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read photo data: %v", err)
		return nil
	}

	// Детектируем лица
	ctx := context.Background()
	faces, err := s.detector.DetectFaces(ctx, photoData)
	if err != nil {
		log.Printf("Failed to detect faces: %v", err)
		return nil
	}

	// Проверяем количество лиц
	facesCount := len(faces)
	if facesCount < s.config.MinFaces || facesCount > s.config.MaxFaces {
		return nil
	}

	// Декодируем изображение для разметки
	img, _, err := imgproc.DecodeImage(photoData)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return nil
	}

	// Отмечаем лица на фото
	markedImg, err := s.marker(img, faces)
	if err != nil {
		log.Printf("Failed to mark faces: %v", err)
		return nil
	}

	// Конвертируем обратно в JPEG
	markedPhotoData, err := imgproc.ConvertToJPEG(markedImg)
	if err != nil {
		log.Printf("Failed to encode marked image: %v", err)
		return nil
	}

	// Отправляем фото с разметкой и кнопкой "Создать голосование"
	btnCreate := tele.InlineButton{
		Unique: fmt.Sprintf("create_voting_%d_%d", c.Chat().ID, c.Message().ID),
		Text:   "Создать голосование",
	}

	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{{btnCreate}},
	}

	photoFile := &tele.Photo{
		File:    tele.FromReader(bytes.NewReader(markedPhotoData)),
		Caption: "Найдено лиц: " + strconv.Itoa(facesCount) + ". Создать голосование?",
	}

	_, err = s.bot.Send(c.Chat(), photoFile, markup)
	if err != nil {
		log.Printf("Failed to send marked photo: %v", err)
		return nil
	}

	return nil
}

// handleCallback обрабатывает callback-запросы (нажатия на кнопки).
func (s *Service) handleCallback(c tele.Context) error {
	data := c.Callback().Data

	// Обработка создания голосования
	if len(data) > 14 && data[:14] == "create_voting_" {
		return s.handleCreateVoting(c)
	}

	// Обработка голосования
	if len(data) > 5 && data[:5] == "vote_" {
		return s.handleVote(c)
	}

	return c.Respond()
}

// handleCreateVoting обрабатывает создание голосования.
func (s *Service) handleCreateVoting(c tele.Context) error {
	// Парсим данные из callback
	// Формат: create_voting_{chatID}_{messageID}
	// Упрощенная версия - в реальности нужно парсить chatID и messageID
	// Для MVP используем текущий чат и сообщение

	chatID := c.Chat().ID
	messageID := c.Message().ID

	// Получаем фото из исходного сообщения
	photo := c.Message().Photo
	if photo == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Фото не найдено"})
	}

	// Загружаем фото для детекции
	file, err := s.bot.File(&photo.File)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка загрузки фото"})
	}

	photoData, err := io.ReadAll(file)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка чтения фото"})
	}

	// Детектируем лица
	ctx := context.Background()
	faces, err := s.detector.DetectFaces(ctx, photoData)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка детекции лиц"})
	}

	facesCount := len(faces)
	if facesCount < s.config.MinFaces || facesCount > s.config.MaxFaces {
		return c.Respond(&tele.CallbackResponse{Text: "Неверное количество лиц"})
	}

	// Создаем сессию голосования
	session, err := s.voting.CreateVoting(
		chatID,
		messageID,
		photo.FileID,
		facesCount,
		s.config.VotingDuration,
	)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка создания голосования"})
	}

	// Создаем inline-кнопки для каждого лица
	buttons := make([]tele.InlineButton, 0, facesCount)
	for i := 0; i < facesCount; i++ {
		btn := tele.InlineButton{
			Unique: fmt.Sprintf("vote_%s_%d", session.ID, i),
			Text:   fmt.Sprintf("👤 №%d", i+1),
		}
		buttons = append(buttons, btn)
	}

	markup := &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{buttons},
	}

	// Отправляем сообщение с голосованием
	text := s.config.VotingText + "\n\nВыберите участника:"
	_, err = s.bot.Send(c.Chat(), text, markup)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка отправки голосования"})
	}

	return c.Respond(&tele.CallbackResponse{Text: "Голосование создано!"})
}

// handleVote обрабатывает голосование.
func (s *Service) handleVote(c tele.Context) error {
	// Парсим данные: vote_{sessionID}_{faceIndex}
	data := c.Callback().Data
	parts := strings.Split(data, "_")
	if len(parts) < 3 {
		return c.Respond(&tele.CallbackResponse{Text: "Неверный формат данных"})
	}

	sessionID := strings.Join(parts[1:len(parts)-1], "_") // sessionID может содержать подчеркивания
	faceIndex, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Неверный индекс лица"})
	}

	// Получаем сессию
	session, exists := s.voting.Store().GetSession(sessionID)
	if !exists {
		return c.Respond(&tele.CallbackResponse{Text: "Сессия голосования не найдена"})
	}

	// Регистрируем голос
	userID := int64(c.Sender().ID)
	success, err := s.voting.Vote(session.ID, userID, faceIndex)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка регистрации голоса"})
	}

	if !success {
		return c.Respond(&tele.CallbackResponse{Text: "Вы уже проголосовали!"})
	}

	return c.Respond(&tele.CallbackResponse{Text: "Ваш голос засчитан!"})
}

// publishResults публикует результаты голосования.
func (s *Service) publishResults(session *voting.Session) error {
	results := session.GetResults()
	winnerIndex, winnerVotes := session.GetWinner()

	// Формируем текст результатов
	var resultText strings.Builder
	resultText.WriteString("🏆 Результаты голосования\n\n")

	// Сортируем результаты по количеству голосов (убывание)
	type resultItem struct {
		faceIndex int
		votes     int
	}
	items := make([]resultItem, 0, len(results))
	for idx, votes := range results {
		items = append(items, resultItem{faceIndex: idx, votes: votes})
	}

	// Простая сортировка по убыванию голосов
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].votes < items[j].votes {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Выводим результаты
	for _, item := range items {
		if item.votes > 0 {
			resultText.WriteString(fmt.Sprintf("№%d — %d голосов\n", item.faceIndex+1, item.votes))
		}
	}

	resultText.WriteString(fmt.Sprintf("\nПобедитель: №%d 🎉 (%d голосов)", winnerIndex+1, winnerVotes))

	// Отправляем результаты в чат
	chat := &tele.Chat{ID: session.ChatID}
	_, err := s.bot.Send(chat, resultText.String())
	return err
}


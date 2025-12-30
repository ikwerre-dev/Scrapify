package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"scrapify/pkg/models"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	ctx    context.Context
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiService{client: client, ctx: ctx}, nil
}

func (s *GeminiService) TranscribeAudio(audioPath string) ([]models.TimelineEntry, error) {
	model := s.client.GenerativeModel("gemini-2.5-flash-lite")

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, err
	}

	prompt := []genai.Part{
		genai.Text("Transcribe this audio file into a JSON array of objects with 'timestamp' (MM:SS) and 'event' (the transcribed text) fields. Return ONLY the JSON array."),
		genai.Blob{MIMEType: "audio/mpeg", Data: audioData},
	}

	resp, err := model.GenerateContent(s.ctx, prompt...)
	if err != nil {
		return nil, err
	}

	var transcription []models.TimelineEntry
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				cleanJSON := strings.TrimSpace(string(txt))
				cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
				cleanJSON = strings.TrimSuffix(cleanJSON, "```")
				err := json.Unmarshal([]byte(cleanJSON), &transcription)
				if err != nil {
					return nil, fmt.Errorf("failed to parse transcription JSON: %v, content: %s", err, cleanJSON)
				}
			}
		}
	}

	return transcription, nil
}

func (s *GeminiService) GenerateStudyGuide(transcription []models.TimelineEntry, gridImagePath string) (*models.StudyGuide, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")

	gridData, err := os.ReadFile(gridImagePath)
	if err != nil {
		return nil, err
	}

	transcriptionJSON, _ := json.Marshal(transcription)

	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Based on the following transcription and the attached image grid (which contains snapshots of the video), generate a comprehensive study guide in JSON format. The JSON should have 'title', 'summary', 'key_points' (array), 'glossary' (array of {term: definition}), and 'timeline' (array of {timestamp: string, event: string}) fields. Transcription: %s", string(transcriptionJSON))),
		genai.Blob{MIMEType: "image/jpeg", Data: gridData},
	}

	resp, err := model.GenerateContent(s.ctx, prompt...)
	if err != nil {
		return nil, err
	}

	var guide models.StudyGuide
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				cleanJSON := strings.TrimSpace(string(txt))
				cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
				cleanJSON = strings.TrimSuffix(cleanJSON, "```")
				err := json.Unmarshal([]byte(cleanJSON), &guide)
				if err != nil {
					return nil, fmt.Errorf("failed to parse study guide JSON: %v, content: %s", err, cleanJSON)
				}
			}
		}
	}

	return &guide, nil
}

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

func (s *GeminiService) GenerateStudyGuide(transcription []models.TimelineEntry, gridImagePaths []string) (*models.StudyGuide, error) {
	model := s.client.GenerativeModel("gemini-2.5-flash-lite")

	transcriptionJSON, _ := json.MarshalIndent(transcription, "", "  ")

	prompt := []genai.Part{
		genai.Text(fmt.Sprintf(`As an advanced educational assistant, analyze the attached video frames and the provided transcription to create a high-quality, comprehensive study guide.

TRANSCRIPTION & TIMELINE:
%s

INSTRUCTIONS:
1. Use the images to understand the visual context (slides, demonstrations, facial expressions).
2. Synthesize the text and visuals into a structured study guide.
3. Return the result strictly in JSON format with these fields:
   - title: A descriptive and engaging title.
   - summary: A 2-3 paragraph high-level overview.
   - key_points: A list of the most important takeaways.
   - glossary: A list of technical terms or concepts with definitions.
   - timeline: A detailed chronological breakdown of events/topics.
`, string(transcriptionJSON))),
	}

	// Add all grid images to the prompt
	for _, path := range gridImagePaths {
		gridData, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read grid image %s: %v", path, err)
		}
		prompt = append(prompt, genai.Blob{MIMEType: "image/jpeg", Data: gridData})
	}

	resp, err := model.GenerateContent(s.ctx, prompt...)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %v", err)
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

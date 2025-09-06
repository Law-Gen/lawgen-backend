package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TranslateResponse struct {
	TranslatedText string `json:"translated_text"`
}

// TranslateText calls the translation API and returns the translated text.
func TranslateText(apiURL, text, targetLang string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"text":        text,
		"target_lang": targetLang,
	})
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", err
	}
	var result TranslateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.TranslatedText, nil
}

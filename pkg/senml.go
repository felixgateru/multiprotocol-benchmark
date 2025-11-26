package pkg

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SenMLRecord struct {
	BaseName    string  `json:"bn,omitempty"`
	BaseTime    float64 `json:"bt,omitempty"`
	BaseUnit    string  `json:"bu,omitempty"`
	BaseVersion int     `json:"bver,omitempty"`
	Name        string  `json:"n,omitempty"`
	Unit        string  `json:"u,omitempty"`
	Value       *string `json:"v,omitempty"`
	StringValue string  `json:"vs,omitempty"`
	BoolValue   *bool   `json:"vb,omitempty"`
	Sum         float64 `json:"s,omitempty"`
	Time        float64 `json:"t,omitempty"`
	UpdateTime  float64 `json:"ut,omitempty"`
}

func CreateSenMLMessage(baseName string, msgNum int, targetSize int) (string, error) {
	timestamp := float64(time.Now().UnixNano()) / 1e9

	paddingValue := ""
	if targetSize > 0 {
		// Create initial record to see base size
		initialRecord := []SenMLRecord{
			{
				BaseName: baseName,
				BaseTime: timestamp,
				Name:     "message",
				Value:    stringPtr("X"),
				Time:     0,
			},
		}

		initialJSON, err := json.Marshal(initialRecord)
		if err != nil {
			return "", fmt.Errorf("failed to marshal initial SenML: %w", err)
		}

		baseSize := len(initialJSON)

		// Calculate padding needed
		if targetSize > baseSize {
			paddingNeeded := targetSize - baseSize + 1 // +1 for the 'X' we already have
			paddingValue = strings.Repeat("X", paddingNeeded)
		} else {
			paddingValue = "X"
		}
	} else {
		paddingValue = fmt.Sprintf("msg_%d", msgNum)
	}

	// Create final record with padding
	record := []SenMLRecord{
		{
			BaseName: baseName,
			BaseTime: timestamp,
			Name:     "message",
			Value:    stringPtr(paddingValue),
			Time:     0,
		},
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SenML: %w", err)
	}

	return string(jsonData), nil
}

func ParseSenMLMessage(payload string) ([]SenMLRecord, error) {
	var records []SenMLRecord
	err := json.Unmarshal([]byte(payload), &records)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SenML: %w", err)
	}
	return records, nil
}

func ExtractMessageNumber(records []SenMLRecord) (int, error) {
	if len(records) == 0 {
		return 0, fmt.Errorf("no records in SenML message")
	}

	baseName := records[0].BaseName
	var msgNum int
	_, err := fmt.Sscanf(baseName, "msg_%d", &msgNum)
	if err == nil {
		return msgNum, nil
	}

	if records[0].Value != nil {
		_, err := fmt.Sscanf(*records[0].Value, "msg_%d", &msgNum)
		if err == nil {
			return msgNum, nil
		}
	}

	return 0, fmt.Errorf("could not extract message number from SenML")
}

func stringPtr(s string) *string {
	return &s
}

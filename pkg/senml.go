package pkg

import (
	"encoding/json"
	"fmt"
	"time"
)

type SenMLRecord struct {
	BaseName    string   `json:"bn,omitempty"`
	BaseTime    float64  `json:"bt,omitempty"`
	BaseUnit    string   `json:"bu,omitempty"`
	BaseVersion int      `json:"bver,omitempty"`
	Name        string   `json:"n,omitempty"`
	Unit        string   `json:"u,omitempty"`
	Value       *float64 `json:"v,omitempty"`
	StringValue string   `json:"vs,omitempty"`
	BoolValue   *bool    `json:"vb,omitempty"`
	Sum         float64  `json:"s,omitempty"`
	Time        float64  `json:"t,omitempty"`
	UpdateTime  float64  `json:"ut,omitempty"`
}

func CreateSenMLMessage(baseName string, msgNum int, targetSize int) (string, error) {
	timestamp := float64(time.Now().UnixNano()) / 1e9

	// Use message number as a simulated sensor value
	sensorValue := float64(msgNum)

	// If targetSize is specified, we need to pad with additional records
	var records []SenMLRecord

	if targetSize > 0 {
		// Create initial record to see base size
		initialRecord := []SenMLRecord{
			{
				BaseName: baseName,
				BaseTime: timestamp,
				Name:     "sensor",
				Unit:     "count",
				Value:    floatPtr(sensorValue),
				Time:     0,
			},
		}

		initialJSON, err := json.Marshal(initialRecord)
		if err != nil {
			return "", fmt.Errorf("failed to marshal initial SenML: %w", err)
		}

		baseSize := len(initialJSON)

		// Add padding records if needed to reach target size
		records = initialRecord
		if targetSize > baseSize {
			// Add extra records with padding data to reach target size
			paddingRecordsNeeded := (targetSize - baseSize) / 50 // Approximate size per record
			for i := 0; i < paddingRecordsNeeded; i++ {
				records = append(records, SenMLRecord{
					Name:  fmt.Sprintf("pad_%d", i),
					Value: floatPtr(0.0),
				})
			}
		}
	} else {
		// Simple record without padding
		records = []SenMLRecord{
			{
				BaseName: baseName,
				BaseTime: timestamp,
				Name:     "sensor",
				Unit:     "count",
				Value:    floatPtr(sensorValue),
				Time:     0,
			},
		}
	}

	jsonData, err := json.Marshal(records)
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

	// Try to extract from base name
	baseName := records[0].BaseName
	var msgNum int
	_, err := fmt.Sscanf(baseName, "msg_%d", &msgNum)
	if err == nil {
		return msgNum, nil
	}

	// Try to extract from numeric value (sensor reading)
	if records[0].Value != nil {
		return int(*records[0].Value), nil
	}

	return 0, fmt.Errorf("could not extract message number from SenML")
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

package types

import "time"

type DER struct {
	DerID             string       `json:"derId"`
	DeviceID          string       `json:"deviceId"`
	Type              string       `json:"type"`
	IsOnline          bool         `json:"isOnline"`
	Timestamp         NillableTime `json:"Timestamp"`
	CurrentOutput     float64      `json:"currentOutput"`
	Units             string       `json:"units"`
	ProjectID         string       `json:"projectId"`
	UtilityID         string       `json:"utilityId"`
	IsStandalone      bool         `json:"isStandalone"`
	ConnectionStartAt NillableTime `json:"connectionStartAt"`
	CurrentSoc        float64      `json:"currentSoc"`
}

type RealTimeDERData struct {
	ID                string       `bigquery:"id" json:"id"`
	DerID             string       `bigquery:"der_id" json:"der_id"`
	DeviceID          string       `bigquery:"device_id" json:"device_id"`
	Timestamp         NillableTime `bigquery:"timestamp" json:"timestamp"`
	Type              string       `bigquery:"type" json:"type"`
	CurrentOutput     float64      `bigquery:"current_output" json:"current_output"`
	Units             string       `bigquery:"units" json:"units"`
	ProjectID         string       `bigquery:"project_id" json:"project_id"`
	IsOnline          bool         `bigquery:"is_online" json:"is_online"`
	IsStandalone      bool         `bigquery:"is_standalone" json:"is_standalone"`
	ConnectionStartAt NillableTime `bigquery:"connection_start_at" json:"connection_start_at"`
	CurrentSoc        float64      `bigquery:"current_soc" json:"current_soc"`
}

type AverageOutput struct {
	ProjectID     string    `bigquery:"project_id" json:"project_id"`
	AverageOutput float64   `bigquery:"average_output" json:"average_output"`
	StartTime     time.Time `bigquery:"start_time" json:"start_time"`
	EndTime       time.Time `bigquery:"end_time" json:"end_time"`
}

package outcome

import "time"

type DER struct {
	DerID             string    `json:"derId"`
	DeviceID          string    `json:"deviceId"`
	Type              string    `json:"type"`
	IsOnline          bool      `json:"isOnline"`
	Timestamp         time.Time `json:"Timestamp"`
	CurrentOutput     float64   `json:"currentOutput"`
	Units             string    `json:"units"`
	ProjectID         string    `json:"projectId"`
	UtilityID         string    `json:"utilityId"`
	IsStandalone      bool      `json:"isStandalone"`
	ConnectionStartAt string    `json:"connectionStartAt"`
	CurrentSoc        int       `json:"currentSoc"`
}

type RealTimeDERData struct {
	ID                string    `bigquery:"id" json:"id"`
	DerID             string    `bigquery:"der_id" json:"der_id"`
	DeviceID          string    `bigquery:"device_id" json:"device_id"`
	Timestamp         time.Time `bigquery:"timestamp" json:"timestamp"`
	CurrentOutput     float64   `bigquery:"current_output" json:"current_output"`
	Units             string    `bigquery:"units" json:"units"`
	ProjectID         string    `bigquery:"project_id" json:"project_id"`
	IsOnline          bool      `bigquery:"is_online" json:"is_online"`
	IsStandalone      bool      `bigquery:"is_standalone" json:"is_standalone"`
	ConnectionStartAt string    `bigquery:"connection_start_at" json:"connection_start_at"`
	CurrentSoc        float64   `bigquery:"current_soc" json:"current_soc"`
}

type AggregateAverageOutput struct {
	ProjectID     string    `bigquery:"project_id" json:"project_id"`
	AverageOutput float64   `bigquery:"average_output" json:"average_output"`
	Timestamp     time.Time `bigquery:"timestamp" json:"timestamp"`
}

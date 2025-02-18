package types

import "time"

type DER struct {
	DerID                 string    `bigquery:"der_id" json:"der_id"`
	IsOnline              bool      `bigquery:"is_online" json:"is_online"`
	Timestamp             time.Time `bigquery:"timestamp" json:"timestamp"`
	CurrentOutput         float64   `bigquery:"current_output" json:"current_output"`
	PowerMeterMeasurement float64   `bigquery:"power_meter_measurement" json:"power_meter_measurement"`
	Baseline              float64   `bigquery:"baseline" json:"baseline"`
	ContractThreshold     float64   `bigquery:"contract_threshold" json:"contract_threshold"`
	Units                 string    `bigquery:"units" json:"units"`
	ProjectID             string    `bigquery:"project_id" json:"project_id"`
	IsStandalone          bool      `bigquery:"is_standalone" json:"is_standalone"`
	ConnectionStartAt     time.Time `bigquery:"connection_start_at" json:"connection_start_at"`
	CurrentSoc            float64   `bigquery:"current_soc" json:"current_soc"`
	Type                  string    `bigquery:"type" json:"type"`
	NameplateCapacity     float64   `bigquery:"nameplate_capacity" json:"nameplate_capacity"`
}

type RealTimeDERData struct {
	ID string `bigquery:"id" json:"id"`
	DER
}

type AverageOutput struct {
	ProjectID         string    `bigquery:"project_id" json:"project_id"`
	AverageOutput     float64   `bigquery:"average_output" json:"average_output"`
	Baseline          float64   `bigquery:"baseline" json:"baseline"`
	ContractThreshold float64   `bigquery:"contract_threshold" json:"contract_threshold"`
	StartTime         time.Time `bigquery:"start_time" json:"start_time"`
	EndTime           time.Time `bigquery:"end_time" json:"end_time"`
}

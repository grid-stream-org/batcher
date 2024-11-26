package models

import "time"

type DER struct {
	DerID             string    `json:"derId"`
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

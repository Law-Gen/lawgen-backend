package domain

import (
	"errors"
	"strings"
)

type LegalEntity struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`

	// --- Core Identification & Status ---
	EntityType          string `json:"entity_type" bson:"entity_type"`                     // e.g., "GOVERNMENT_AGENCY", "PRIVATE_LAW_FIRM", "NGO"
	DateOfEstablishment string `json:"date_of_establishment" bson:"date_of_establishment"` // e.g., "1998-05-20"
	Status              string `json:"status" bson:"status"`                               // e.g., "ACTIVE", "INACTIVE"

	// --- Contact Information ---
	Phone   []string `json:"phone" bson:"phone"` // Can have multiple phone numbers
	Email   []string `json:"email" bson:"email"` // Can have multiple emails
	Website string   `json:"website" bson:"website"`

	// --- Detailed Location ---
	City          string `json:"city" bson:"city"`
	SubCity       string `json:"sub_city" bson:"sub_city"`
	Woreda        string `json:"woreda" bson:"woreda"`
	StreetAddress string `json:"street_address" bson:"street_address"`

	// --- Purpose & Services ---
	Description     string   `json:"description" bson:"description"`           // A summary of the entity's purpose
	ServicesOffered []string `json:"services_offered" bson:"services_offered"` // For a law firm, this is "Areas of Practice"
	Jurisdiction    string   `json:"jurisdiction" bson:"jurisdiction"`         // Relevant for government bodies

	// --- Operational Details ---
	WorkingHours  string `json:"working_hours" bson:"working_hours"`
	ContactPerson string `json:"contact_person" bson:"contact_person"`
}

func (e *LegalEntity) IsValid() error {
	if strings.TrimSpace(e.Name) == "" {
		return errors.New("name is a required field")
	}

	if strings.TrimSpace(e.EntityType) == "" {
		return errors.New("entity_type is a required field")
	}

	if len(e.Phone) == 0 || strings.TrimSpace(e.Phone[0]) == "" {
		return errors.New("at least one phone number is required")
	}

	if strings.TrimSpace(e.City) == "" {
		return errors.New("city is a required field")
	}

	return nil 
}


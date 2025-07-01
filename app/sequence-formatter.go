package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func (s *Server) GenerateNextSequence(tx *sql.Tx, companyID int, module string, subType string) (string, error) {
	var sequenceJSON []byte
	err := tx.QueryRow("SELECT sequences FROM companies_setttings WHERE company_id = $1 FOR UPDATE", companyID).Scan(&sequenceJSON)
	if err != nil {
		return "", err
	}

	var sequences map[string]any
	if err := json.Unmarshal(sequenceJSON, &sequences); err != nil {
		return "", err
	}

	// Drill into the config depending on module and subtype
	var cfg map[string]any
	switch typed := sequences[module].(type) {
	case map[string]any:
		if subType != "" {
			cfg, _ = typed[subType].(map[string]any)
		} else {
			cfg = typed
		}
	}

	if cfg == nil {
		return "", fmt.Errorf("sequence config not found for module: %s", module)
	}

	prefix := cfg["prefix"].(string)
	next := int(cfg["next"].(float64))
	padding := int(cfg["padding"].(float64))

	seqStr := fmt.Sprintf("%s%0*d", prefix, padding, next)

	// Increment and store back
	cfg["next"] = next + 1

	if subType != "" {
		sequences[module].(map[string]any)[subType] = cfg
	} else {
		sequences[module] = cfg
	}

	updatedJSON, err := json.Marshal(sequences)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`UPDATE business_settings SET sequences = $1, updated_at = now() WHERE business_id = $2`, updatedJSON, companyID)
	if err != nil {
		return "", err
	}

	return seqStr, nil
}

package app

import (
	"database/sql"
	"fmt"
	"strings"
)

type SequenceInfo struct {
	Next    int
	Prefix  string
	Padding int
	Code    string // formatted code
}

// Helper to convert "invoice.credit" → {invoice,credit,next}
func buildJSONPath(path string) string {
	segments := strings.Split(path, ".")
	segments = append(segments, "next")
	return "{" + strings.Join(segments, ",") + "}"
}

func GetNextSequence(tx *sql.Tx, companyID int, jsonPath string) (*SequenceInfo, error) {
	// Build the array path for jsonb_set and json access
	pathArray := buildJSONPath(jsonPath)

	query := fmt.Sprintf(`
    WITH current AS (
        SELECT
            company_id,
            (sequences#>>'%s')::int        AS seq,
            sequences#>>'{%s,prefix}'      AS prefix,
            (sequences#>>'{%s,padding}')::int AS padding
        FROM companies_settings
        WHERE company_id = $1
        -- Stays raw: the lock is taken inside a CTE that a jsonb_set UPDATE then
        -- reads from. playsql can lock a Builder's own SELECT, not a CTE arm.
        FOR UPDATE
    ),
    updated AS (
        UPDATE companies_settings
        SET sequences = jsonb_set(
            sequences,
            '%s',
            to_jsonb(current.seq + 1),
            false
        )
        FROM current
        WHERE companies_settings.company_id = current.company_id
        RETURNING current.seq, current.prefix, current.padding
    )
    SELECT
        prefix || lpad(seq::text, padding, '0') AS code,
        seq, prefix, padding
    FROM updated;
`, pathArray, strings.Join(strings.Split(jsonPath, "."), ","), strings.Join(strings.Split(jsonPath, "."), ","), pathArray)

	var info SequenceInfo
	if err := tx.QueryRow(query, companyID).Scan(&info.Code, &info.Next, &info.Prefix, &info.Padding); err != nil {
		return nil, err
	}

	return &info, nil
}

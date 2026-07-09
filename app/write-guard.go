package app

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrRecordNotFound is returned when a write that targets exactly one row — an
// update, a soft delete, a status toggle — matches none.
//
// Every such statement in this codebase is scoped by company_id, so a zero-row
// result means the id or uuid names a record that does not exist, or belongs to
// another company. Neither is a success.
var ErrRecordNotFound = errors.New("record not found")

// mustAffectRow turns a zero-row write into an error.
//
// A statement like `UPDATE customers SET ... WHERE company_id = $1 AND id = $2`
// affects no rows when the id is wrong or belongs to another tenant, and
// database/sql reports that as a successful Exec. The handler above then flashes
// "updated" and redirects, having changed nothing. Three of the bugs found while
// migrating this codebase to playsql were exactly this shape:
//
//   - updateSequences and friends compared company_id against a NULL subquery
//     (fixed in a4662d0);
//   - failUpload passed its arguments transposed, matching `WHERE id = <the error
//     message>` (fixed in 0f214de);
//   - attachItemUnit's ON CONFLICT could never fire (fixed in 8515537).
//
// Only two places in the codebase checked RowsAffected before this helper existed:
// updateCustomerAmountDue and updateVendorAmountPayable, both balance updates.
//
// Callers that legitimately expect zero rows (a conditional write, a bulk update)
// must not use this — check the count themselves.
func mustAffectRow(res sql.Result, err error, subject string) error {
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		// The driver does not report the count. Do not invent a failure.
		return nil
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", subject, ErrRecordNotFound)
	}
	return nil
}

// mustAffectRows is the playsql equivalent: Builder.Update and Builder.Delete
// return the affected count directly.
func mustAffectRows(affected int64, err error, subject string) error {
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", subject, ErrRecordNotFound)
	}
	return nil
}

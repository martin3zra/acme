package app

import (
	"errors"
	"testing"
	"time"

	"github.com/martin3zra/playsql"
)

// Company reads and the companies_settings statements, converted to playsql. Every
// settings statement used to inline
// `WHERE company_id = (SELECT id FROM companies WHERE account_id = $1 AND uuid = $2)`;
// the lookup is now hoisted into resolveCompanyID.
//
// storeCompany and linkCompanyDefaultSequences are already exercised by every test
// through mkAccountCompany, so these cover the reads and the settings round-trips.

func TestFindCompanies(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s) // a second account

	companies, err := s.findCompanies(f.ctx)
	is.NoErr(err)
	is.Equal(len(companies), 1) // scoped to the account
	is.Equal(companies[0].ID, f.company.ID)
	is.True(companies[0].Name != "", "name should be scanned")
	is.True(companies[0].ID != other.company.ID, "another account's company must not appear")
}

func TestFindCompanyByUUIDAndID(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	uuid := companyUUIDOf(t, f)

	byUUID, err := s.findCompanyByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(byUUID.ID, f.company.ID)
	is.Equal(byUUID.Identifier, "131000000")
	is.Equal(byUUID.City, "Santo Domingo")

	byID, err := s.findCompanyByID(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(byID.UUID, uuid)

	// findCompanyByUUID is account-scoped; findCompanyByID is not.
	_, err = s.findCompanyByUUID(f.ctx, companyUUIDOf(t, other))
	is.Err(err, "another account's company must not resolve by uuid")

	foreign, err := s.findCompanyByID(f.ctx, other.company.ID)
	is.NoErr(err)
	is.Equal(foreign.ID, other.company.ID)
}

// TestFindSequences: linkCompanyDefaultSequences seeds the row through playsql's
// Upsert, and the jsonb blob decodes into CompanySequence.
func TestFindSequences(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	seq, err := s.findSequences(f.ctx, companyUUIDOf(t, f))
	is.NoErr(err)
	is.Equal(seq.Sequence.Invoice.Cash.Prefix, "INV-CO-")
	is.Equal(seq.Sequence.Invoice.Credit.Prefix, "INV-CRE-")
	is.Equal(seq.Sequence.Customer.Prefix, "CUST-")
	is.Equal(seq.Sequence.Customer.Padding, 6)
	is.True(!seq.UpdatedAt.IsZero(), "updated_at should be scanned")
}

// TestUpdateSequences round-trips the blob and confirms playsql stamps updated_at
// where the raw statement said `updated_at = now()`.
func TestUpdateSequences(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	uuid := companyUUIDOf(t, f)

	before, err := s.findSequences(f.ctx, uuid)
	is.NoErr(err)

	time.Sleep(2 * time.Millisecond)
	form := &SequenceForm{}
	form.Customer = SequenceConfig{Prefix: "CLI-", Next: 42, Padding: 3}
	is.NoErr(s.updateSequences(f.ctx, uuid, form))

	after, err := s.findSequences(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(after.Sequence.Customer.Prefix, "CLI-")
	is.Equal(after.Sequence.Customer.Next, 42)
	is.True(after.UpdatedAt.After(before.UpdatedAt), "updateSequences should bump updated_at")
}

func TestRedirectPreferences(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	uuid := companyUUIDOf(t, f)

	prefs, err := s.findRedirectPreferences(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(string(prefs.Redirect.Invoice), string(RedirectPreference.Stay))
	is.Equal(string(prefs.Redirect.Customer), string(RedirectPreference.List))

	is.NoErr(s.updateRedirectPreferences(f.ctx, uuid, &RedirectPreferencesForm{
		Invoice: string(RedirectPreference.Detail), Customer: string(RedirectPreference.Stay),
	}))

	updated, err := s.findRedirectPreferences(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(string(updated.Redirect.Invoice), string(RedirectPreference.Detail))
	is.Equal(string(updated.Redirect.Customer), string(RedirectPreference.Stay))
}

// TestHandlesVariants: the flag defaults false, toggles, and a company with no
// settings row reports false rather than erroring.
//
// handles_variants is NOT NULL with a false default, so the old query's
// COALESCE(handles_variants, false) never had anything to coalesce — the missing-row
// guard was doing all the work.
func TestHandlesVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	uuid := companyUUIDOf(t, f)

	enabled, err := s.findHandlesVariants(f.ctx, uuid)
	is.NoErr(err)
	is.True(!enabled, "the flag defaults to false")

	is.NoErr(s.updateHandlesVariants(f.ctx, uuid, true))

	enabled, err = s.findHandlesVariants(f.ctx, uuid)
	is.NoErr(err)
	is.True(enabled, "the flag should toggle on")

	byID, err := s.handlesVariantsByCompanyID(f.ctx, f.company.ID)
	is.NoErr(err)
	is.True(byID, "reading by company id agrees")

	// A company with no settings row at all reports false, not an error.
	_, err = s.db.Exec(`DELETE FROM companies_settings WHERE company_id = $1`, f.company.ID)
	is.NoErr(err)
	byID, err = s.handlesVariantsByCompanyID(f.ctx, f.company.ID)
	is.NoErr(err)
	is.True(!byID, "a missing settings row reports false")
}

// TestCompanySettings_UnknownUUID pins a deliberate behaviour change.
//
// The old statements compared company_id against a scalar subquery that yielded
// NULL for an unknown uuid. `company_id = NULL` matches nothing, so the reads
// returned sql.ErrNoRows but the UPDATEs affected zero rows and reported success —
// a settings write against a bogus or another account's company looked like it
// worked. resolveCompanyID now fails first, so both report not-found.
func TestCompanySettings_UnknownUUID(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	unknown := "00000000-0000-0000-0000-000000000000"
	foreign := companyUUIDOf(t, other) // real, but belongs to another account

	for _, uuid := range []string{unknown, foreign} {
		_, err := s.findSequences(f.ctx, uuid)
		is.True(errors.Is(err, playsql.ErrNotFound), "findSequences should not find it")

		err = s.updateSequences(f.ctx, uuid, &SequenceForm{})
		is.True(errors.Is(err, playsql.ErrNotFound), "updateSequences must fail, not silently no-op")

		err = s.updateHandlesVariants(f.ctx, uuid, true)
		is.True(errors.Is(err, playsql.ErrNotFound), "updateHandlesVariants must fail, not silently no-op")

		err = s.updateRedirectPreferences(f.ctx, uuid, &RedirectPreferencesForm{})
		is.True(errors.Is(err, playsql.ErrNotFound), "updateRedirectPreferences must fail, not silently no-op")
	}

	// The other account's settings were not touched.
	enabled, err := s.handlesVariantsByCompanyID(f.ctx, other.company.ID)
	is.NoErr(err)
	is.True(!enabled, "the foreign company's flag must be untouched")
}

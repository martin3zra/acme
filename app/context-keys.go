package app

// AccountKey and CompanyKey are context keys for the current tenant (account)
// and the active company. They are application/tenant concepts and therefore
// live in the app, not in the reusable framework.
type AccountKey struct{}

type CompanyKey struct{}

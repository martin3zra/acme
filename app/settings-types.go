package app

import (
	"github.com/martin3zra/forge/support"
)

type RedirectPreferencesValue string

const (
	_STAY   RedirectPreferencesValue = "stay"
	_LIST   RedirectPreferencesValue = "list"
	_DETAIL RedirectPreferencesValue = "detail"
)

type RedirectPreferences struct {
	Invoice  RedirectPreferencesValue `json:"invoice"`
	Estimate RedirectPreferencesValue `json:"estimate"`
	Customer RedirectPreferencesValue `json:"customer"`
	Vendor   RedirectPreferencesValue `json:"vendor"`
	Item     RedirectPreferencesValue `json:"item"`
	Payment  RedirectPreferencesValue `json:"payment"`
	Order    RedirectPreferencesValue `json:"order"`
}

var RedirectPreference = struct {
	Stay   RedirectPreferencesValue
	List   RedirectPreferencesValue
	Detail RedirectPreferencesValue
}{
	Stay:   _STAY,
	List:   _LIST,
	Detail: _DETAIL,
}

type RedirectPreferencesForm struct {
	support.FormRequest
	Invoice  string `json:"invoice"`
	Estimate string `json:"estimate"`
	Customer string `json:"customer"`
	Vendor   string `json:"vendor"`
	Order    string `json:"order"`
	Item     string `json:"item"`
	Payment  string `json:"payment"`
}

func (RedirectPreferencesForm) Rules() map[string]any {
	return map[string]any{
		"invoice":  "required|in:list,detail,stay",
		"estimate": "required|in:list,detail,stay",
		"customer": "required|in:list,detail,stay",
		"vendor":   "required|in:list,detail,stay",
		"order":    "required|in:list,detail,stay",
		"item":     "required|in:list,detail,stay",
		"payment":  "required|in:list,detail,stay",
	}
}

// HandlesVariantsForm toggles the company-level product-variants feature flag.
type HandlesVariantsForm struct {
	support.FormRequest
	Enabled bool `json:"enabled"`
}

func (HandlesVariantsForm) Rules() map[string]any {
	return map[string]any{
		"enabled": "boolean",
	}
}

type StoreUnitForm struct {
	support.FormRequest
	Name    string `json:"name"`
	BaseQty int    `json:"base_qty"`
}

func (StoreUnitForm) Rules() map[string]any {
	return map[string]any{
		"name":     "required|min:1",
		"base_qty": "required|min:1",
	}
}

package app

import (
	"net/http"
	"regexp"

	"github.com/martin3zra/acme/pkg/foundation"
)

func flash(w http.ResponseWriter, name string, value any) {
	foundation.SetFlash(w, name, value)
}

func ensureUUIDIsValid(str string) bool {
	regex := `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
	r := regexp.MustCompile(regex)
	return r.MatchString(str)
}

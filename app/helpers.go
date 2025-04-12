package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/foundation"
)

func flash(w http.ResponseWriter, name string, value any) {
	foundation.SetFlash(w, name, value)
}

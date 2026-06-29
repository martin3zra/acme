package app

import "testing"

func TestMapHeaders_Vendors(t *testing.T) {
	// Mixed case + an unrecognized column that must be ignored.
	headers := []string{"Nombre", "correo", "TELEFONO", "Codigo", "Tipo", "Unknown"}

	got, err := mapHeaders(headers, "vendors")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[int]string{
		0: "name",
		1: "email",
		2: "phone",
		3: "code",
		4: "vendor_type",
	}
	if len(got) != len(want) {
		t.Fatalf("want %d mapped columns, got %d (%v)", len(want), len(got), got)
	}
	for idx, field := range want {
		if got[idx] != field {
			t.Errorf("column %d: want %q, got %q", idx, field, got[idx])
		}
	}
	if _, ok := got[5]; ok {
		t.Error("unrecognized header should not be mapped")
	}
}

func TestMapHeaders_SourcesAreDistinct(t *testing.T) {
	// TIPO maps to vendor_type for vendors, but is not a customer field
	// (customers use TIPO_NCFTP for tax_receipt_id).
	if got, _ := mapHeaders([]string{"TIPO"}, "vendors"); got[0] != "vendor_type" {
		t.Errorf("vendors TIPO: want vendor_type, got %q", got[0])
	}
	if got, _ := mapHeaders([]string{"TIPO"}, "customers"); len(got) != 0 {
		t.Errorf("customers bare TIPO should not map, got %v", got)
	}
	if got, _ := mapHeaders([]string{"TIPO_NCFTP"}, "customers"); got[0] != "tax_receipt_id" {
		t.Errorf("customers TIPO_NCFTP: want tax_receipt_id, got %q", got[0])
	}
}

func TestDetectDelimiter(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  rune
	}{
		{"comma", []string{"a,b,c", "1,2,3"}, ','},
		{"semicolon", []string{"a;b;c", "1;2;3"}, ';'},
		{"tab", []string{"a\tb\tc", "1\t2\t3"}, '\t'},
		{"pipe", []string{"a|b|c", "1|2|3"}, '|'},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DetectDelimiter(c.lines); got != c.want {
				t.Errorf("want %q, got %q", c.want, got)
			}
		})
	}
}

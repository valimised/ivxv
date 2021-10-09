package main

import (
	"testing"

	"ivxv.ee/errors"
	"ivxv.ee/storage"
)

func testDistrictsLookupFormatError(t *testing.T, id string) {
	t.Helper()
	list := districtlist{Districts: map[string]district{id: {}}}

	_, err := districtsLookup(list)
	if errors.CausedBy(err, new(DistrictIDFormatError)) == nil {
		t.Error("unexpected success with malformed district code", id)
	}
}

func TestDistrictsLookupFormatError(t *testing.T) {
	testDistrictsLookupFormatError(t, "0000")
	testDistrictsLookupFormatError(t, "0000.")
	testDistrictsLookupFormatError(t, ".0000")
	testDistrictsLookupFormatError(t, "0000.0.0")
}

func TestDistrictsLookupMultipleCodes(t *testing.T) {
	list := districtlist{
		Districts: map[string]district{
			"0001.1": {
				Parish: []string{"1111"},
			},
			"0002.2": {
				Parish: []string{"1111"},
			},
		},
	}

	_, err := districtsLookup(list)
	if errors.CausedBy(err, new(ParishWithMultipleDistrictCodesError)) == nil {
		t.Error("unexpected success with multiple district codes")
	}
}

func assertDistricts(t *testing.T, districts map[string][]byte,
	adminCode, districtNr, districtID string) {

	t.Helper()
	encoded := storage.EncodeAdminDistrict(adminCode, districtNr)
	id, ok := districts[string(encoded)]
	if !ok {
		t.Error("unexpected error: administrative unit code", adminCode,
			"and district number", districtNr, "not in districts")
	}
	if string(id) != districtID {
		t.Errorf("unexpected district identifier for %s and %s: got %s, want %s",
			adminCode, districtNr, id, districtID)
	}
}

func TestDistrictsLookup(t *testing.T) {
	list := districtlist{
		Districts: map[string]district{
			// One-to-one match.
			"0001.1": {
				Parish: []string{"1111"},
			},

			// District with multiple parishes.
			"0002.1": {
				Parish: []string{"2222", "3333"},
			},

			// Parish with multiple districts.
			"0004.1": {
				Parish: []string{"4444"},
			},
			"0004.2": {
				Parish: []string{"4444"},
			},
		},
	}

	districts, err := districtsLookup(list)
	if err != nil {
		t.Fatal(err)
	}

	assertDistricts(t, districts, "1111", "1", "0001.1")
	assertDistricts(t, districts, "2222", "1", "0002.1")
	assertDistricts(t, districts, "3333", "1", "0002.1")
	assertDistricts(t, districts, "4444", "1", "0004.1")
	assertDistricts(t, districts, "4444", "2", "0004.2")
}

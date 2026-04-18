package db

import (
	"testing"
	"time"
)

func TestNormalizeTimeZone(t *testing.T) {
	if NormalizeTimeZone("") != "UTC" {
		t.Fatal()
	}
	if NormalizeTimeZone("  Asia/Shanghai  ") != "Asia/Shanghai" {
		t.Fatal()
	}
}

func TestLocationForTimeZone(t *testing.T) {
	loc, err := LocationForTimeZone("UTC")
	if err != nil || loc != time.UTC {
		t.Fatal(err)
	}
	loc, err = LocationForTimeZone("+08:00")
	if err != nil {
		t.Fatal(err)
	}
	utcNoon := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	if utcNoon.In(loc).Hour() != 20 {
		t.Fatalf("want +8 offset, got local hour %d", utcNoon.In(loc).Hour())
	}
	_, err = LocationForTimeZone("+99:00")
	if err == nil {
		t.Fatal("expect error")
	}
}

func TestMysqlSetTimeZoneValue(t *testing.T) {
	if mysqlSetTimeZoneValue("UTC") != "+00:00" {
		t.Fatal()
	}
	if mysqlSetTimeZoneValue("+05:30") != "+05:30" {
		t.Fatal()
	}
}

func TestFormatMySQLTimeZoneOffset(t *testing.T) {
	if formatMySQLTimeZoneOffset(0) != "+00:00" {
		t.Fatal()
	}
	if formatMySQLTimeZoneOffset(-3600) != "-01:00" {
		t.Fatal()
	}
	if formatMySQLTimeZoneOffset(5400) != "+01:30" {
		t.Fatal()
	}
}

func TestDbTimeZoneIsUTCEquivalent(t *testing.T) {
	if !dbTimeZoneIsUTCEquivalent("+00:00") || !dbTimeZoneIsUTCEquivalent("-00:00") {
		t.Fatal()
	}
	if dbTimeZoneIsUTCEquivalent("EST") {
		t.Fatal()
	}
}

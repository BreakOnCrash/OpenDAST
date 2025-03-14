package jsonp

import "testing"

func TestAuditJSONPHijacking(t *testing.T) {
	AuditJSONPHijacking("http://localhost:8080/jsonp?callback=func")
}

package eventcounter

import "testing"

func TestConstants(t *testing.T) {
	if EventCreated != "created" {
		t.Errorf("EventCreated incorreto: %s", EventCreated)
	}
	
	if EventUpdated != "updated" {
		t.Errorf("EventUpdated incorreto: %s", EventUpdated)
	}
	
	if EventDeleted != "deleted" {
		t.Errorf("EventDeleted incorreto: %s", EventDeleted)
	}
}

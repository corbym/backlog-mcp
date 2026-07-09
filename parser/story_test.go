package parser

import "testing"

func TestParseACText_PlainSequentialID(t *testing.T) {
	id, text := parseACText("AC-STORY-159-1: Some criterion text")
	if id != "AC-STORY-159-1" {
		t.Errorf("id: got %q, want %q", id, "AC-STORY-159-1")
	}
	if text != "Some criterion text" {
		t.Errorf("text: got %q, want %q", text, "Some criterion text")
	}
}

func TestParseACText_PlainSequentialID_MultiDigit(t *testing.T) {
	id, text := parseACText("AC-STORY-159-12: Twelfth criterion")
	if id != "AC-STORY-159-12" {
		t.Errorf("id: got %q, want %q", id, "AC-STORY-159-12")
	}
	if text != "Twelfth criterion" {
		t.Errorf("text: got %q, want %q", text, "Twelfth criterion")
	}
}

func TestParseACText_HexSuffixedID_StillRecognised(t *testing.T) {
	id, text := parseACText("AC-STORY-042-a3f9b2c1: User can log in")
	if id != "AC-STORY-042-a3f9b2c1" {
		t.Errorf("id: got %q, want %q", id, "AC-STORY-042-a3f9b2c1")
	}
	if text != "User can log in" {
		t.Errorf("text: got %q, want %q", text, "User can log in")
	}
}

func TestParseACText_NoID_ReturnsFullTextUnchanged(t *testing.T) {
	id, text := parseACText("Plain criterion with no ID prefix")
	if id != "" {
		t.Errorf("id: got %q, want empty", id)
	}
	if text != "Plain criterion with no ID prefix" {
		t.Errorf("text: got %q, want unchanged full text", text)
	}
}

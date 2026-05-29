package store

import (
	"testing"

	"life-sim/backend/model"
)

func TestValidateGameChoicesForAge_childRejectsArmy(t *testing.T) {
	opts := []model.GameChoiceOption{
		{ID: "c1", Label: "率军北伐", Description: "统帅三军"},
	}
	if err := ValidateGameChoicesForAge(8, opts); err == nil {
		t.Fatal("expected rejection for child army command")
	}
}

func TestValidateGameChoicesForAge_adultPasses(t *testing.T) {
	opts := []model.GameChoiceOption{
		{ID: "c1", Label: "率军北伐", Description: "统帅三军"},
	}
	if err := ValidateGameChoicesForAge(28, opts); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateGameChoicesForAge_childPlayOk(t *testing.T) {
	opts := []model.GameChoiceOption{
		{ID: "c1", Label: "随父迁居", Description: "离开故里"},
	}
	if err := ValidateGameChoicesForAge(8, opts); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

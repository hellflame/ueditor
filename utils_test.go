package ueditor

import (
	"testing"
)

func TestLowerCamalMarshal(t *testing.T) {
	type linux struct {
		Name string
	}
	if string(lowerCamalMarshal(linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
	if string(lowerCamalMarshal(&linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
}

func TestParseDefault(t *testing.T) {
	type linux struct {
		Name   string   `default:"fine"`
		Age    int      `default:"18"`
		Weight float64  `default:"12.5"`
		Jobs   []string `default:"chef|waiter|teacher"`
	}
	x := &linux{}
	applyDefault(x)
	if x.Name != "fine" || x.Age != 18 || x.Weight < 12 || x.Jobs[1] != "waiter" {
		t.Error("failed to apply")
	}
}

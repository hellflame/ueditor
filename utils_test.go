package ueditor

import (
	"testing"
)

func TestLowerCamalMarshal(t *testing.T) {
	type linux struct {
		Name string
	}
	if string(LowerCamalMarshal(linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
	if string(LowerCamalMarshal(&linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
	tmp := []linux{{"hellflame"}}
	if string(LowerCamalMarshal(tmp))[3] != 'n' {
		t.Error("slice element not lower")
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

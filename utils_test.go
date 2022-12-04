package ueditor

import (
	"testing"
)

func TestLowerCamelMarshal(t *testing.T) {
	type linux struct {
		Name string
	}
	if string(LowerCamelMarshal(linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
	if string(LowerCamelMarshal(&linux{Name: "ok"}))[2] != 'n' {
		t.Error("not lower")
	}
	tmp := []linux{{"hellflame"}}
	if string(LowerCamelMarshal(tmp))[3] != 'n' {
		t.Error("slice element not lower")
	}
}

func TestParseDefault(t *testing.T) {
	type Part struct {
		Parent string `default:"ok"`
	}
	type linux struct {
		Name   string   `default:"fine"`
		Age    int      `default:"18"`
		Weight float64  `default:"12.5"`
		Jobs   []string `default:"chef|waiter|teacher"`
		Part
		extra *Part
	}
	x := &linux{}
	applyDefault(x)
	if x.Name != "fine" || x.Age != 18 || x.Weight < 12 || x.Jobs[1] != "waiter" {
		t.Error("failed to apply")
	}
	if x.Parent != "ok" {
		t.Error("failed to apply for sub struct")
	}
	if x.extra != nil {
		t.Error("pointer should not be applied")
	}
}

func TestIsAllowedFileType(t *testing.T) {
	if !isAllowedFileType("linux.a.img", []string{".img"}) {
		t.Error("failed to judge")
	}
	if isAllowedFileType("ok", []string{}) {
		t.Error("fail to judge")
	}
}

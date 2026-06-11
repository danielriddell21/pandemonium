package hud

import "testing"

func TestOverlayPostAndExpire(t *testing.T) {
	o := New()
	if _, _, ok := o.Active(); ok {
		t.Fatal("new overlay should be inactive")
	}

	o.Post("area cleared", 3, Notice)
	msg, ch, ok := o.Active()
	if !ok || msg != "area cleared" || ch != Notice {
		t.Fatalf("Active = %q,%v,%v; want message active on Notice", msg, ch, ok)
	}

	o.Tick() // 3 -> 2
	o.Tick() // 2 -> 1
	if _, _, ok := o.Active(); !ok {
		t.Fatal("message expired too early")
	}
	o.Tick() // 1 -> 0
	if msg, _, ok := o.Active(); ok || msg != "" {
		t.Fatalf("message should have expired, got %q,%v", msg, ok)
	}
}

func TestOverlayReplace(t *testing.T) {
	o := New()
	o.Post("first", 10, Diagnostic)
	o.Post("second", 2, Notice)
	msg, ch, _ := o.Active()
	if msg != "second" || ch != Notice {
		t.Errorf("Post should replace: got %q on %v, want second on Notice", msg, ch)
	}
}

func TestOverlayPostEmptyOrNonPositiveClears(t *testing.T) {
	o := New()
	o.Post("x", 5, Notice)
	o.Post("", 5, Notice)
	if _, _, ok := o.Active(); ok {
		t.Error("empty text should clear the overlay")
	}
	o.Post("y", 0, Notice)
	if _, _, ok := o.Active(); ok {
		t.Error("non-positive frames should clear the overlay")
	}
}

func TestOverlayTickWhenIdle(t *testing.T) {
	o := New()
	o.Tick() // must not panic or go negative
	if _, _, ok := o.Active(); ok {
		t.Error("idle overlay should stay inactive")
	}
}

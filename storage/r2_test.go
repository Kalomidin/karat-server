package storage

import "testing"

func TestAKeyBelongsOnlyToItsOwner(t *testing.T) {
	store := unavailableStore{}
	own := "items/ada/photo.jpg"
	if !store.BelongsTo("ada", own) {
		t.Fatal("owner key was rejected")
	}
	if store.BelongsTo("bea", own) || store.BelongsTo("ada", "items/ada/../bea/photo.jpg") {
		t.Fatal("foreign key was accepted")
	}
	if _, ok := extensionFor("image/jpeg"); !ok {
		t.Fatal("jpeg")
	}
	if _, ok := extensionFor("image/gif"); ok {
		t.Fatal("gif should be rejected")
	}
}

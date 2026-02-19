package enviar

import (
	"encoding/json"
	"testing"
)

func TestBodyArgs_StringPayload(t *testing.T) {
	args, err := bodyArgs("hello")
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}
	raw, ok := args[BodyKey]
	if !ok {
		t.Fatal("missing body key in args")
	}
	body, ok := raw.(string)
	if !ok {
		t.Fatalf("body is %T, want string", raw)
	}
	if body != `"hello"` {
		t.Errorf("body = %q, want %q", body, `"hello"`)
	}
}

func TestBodyArgs_StructPayload(t *testing.T) {
	type Email struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
	}
	payload := Email{To: "alice@example.com", Subject: "Hi"}

	args, err := bodyArgs(payload)
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}

	body := args[BodyKey].(string)

	var got Email
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.To != "alice@example.com" {
		t.Errorf("To = %q, want %q", got.To, "alice@example.com")
	}
	if got.Subject != "Hi" {
		t.Errorf("Subject = %q, want %q", got.Subject, "Hi")
	}
}

func TestBodyArgs_MapPayload(t *testing.T) {
	payload := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}
	args, err := bodyArgs(payload)
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}

	body := args[BodyKey].(string)

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("key = %v, want %q", got["key"], "value")
	}
}

func TestBodyArgs_NilPayload(t *testing.T) {
	args, err := bodyArgs(nil)
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}
	body := args[BodyKey].(string)
	if body != "null" {
		t.Errorf("body = %q, want %q", body, "null")
	}
}

func TestBodyArgs_UnmarshalablePayload(t *testing.T) {
	// Channels are not JSON-serializable.
	ch := make(chan int)
	_, err := bodyArgs(ch)
	if err == nil {
		t.Fatal("expected error for unmarshalable payload")
	}
}

func TestBodyArgs_SlicePayload(t *testing.T) {
	payload := []string{"a", "b", "c"}
	args, err := bodyArgs(payload)
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}

	body := args[BodyKey].(string)

	var got []string
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("got %v, want [a b c]", got)
	}
}

func TestBodyArgs_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	payload := Person{
		Name:    "Alice",
		Address: Address{City: "Nairobi"},
	}

	args, err := bodyArgs(payload)
	if err != nil {
		t.Fatalf("bodyArgs: %v", err)
	}

	body := args[BodyKey].(string)

	var got Person
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("Name = %q, want %q", got.Name, "Alice")
	}
	if got.Address.City != "Nairobi" {
		t.Errorf("City = %q, want %q", got.Address.City, "Nairobi")
	}
}

func TestBodyKey_Constant(t *testing.T) {
	if BodyKey != "body" {
		t.Errorf("BodyKey = %q, want %q", BodyKey, "body")
	}
}

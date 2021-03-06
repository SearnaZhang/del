package whisperv2
import (
	"bytes"
	"testing"
	"time"
	"github.com/DEL-ORG/del/crypto"
	"github.com/DEL-ORG/del/crypto/ecies"
)
func TestEnvelopeOpen(t *testing.T) {
	payload := []byte("hello world")
	message := NewMessage(payload)
	envelope, err := message.Wrap(DefaultPoW, Options{})
	if err != nil {
		t.Fatalf("failed to wrap message: %v", err)
	}
	opened, err := envelope.Open(nil)
	if err != nil {
		t.Fatalf("failed to open envelope: %v", err)
	}
	if opened.Flags != message.Flags {
		t.Fatalf("flags mismatch: have %d, want %d", opened.Flags, message.Flags)
	}
	if !bytes.Equal(opened.Signature, message.Signature) {
		t.Fatalf("signature mismatch: have 0x%x, want 0x%x", opened.Signature, message.Signature)
	}
	if !bytes.Equal(opened.Payload, message.Payload) {
		t.Fatalf("payload mismatch: have 0x%x, want 0x%x", opened.Payload, message.Payload)
	}
	if opened.Sent.Unix() != message.Sent.Unix() {
		t.Fatalf("send time mismatch: have %v, want %v", opened.Sent, message.Sent)
	}
	if opened.TTL/time.Second != DefaultTTL/time.Second {
		t.Fatalf("message TTL mismatch: have %v, want %v", opened.TTL, DefaultTTL)
	}
	if opened.Hash != envelope.Hash() {
		t.Fatalf("message hash mismatch: have 0x%x, want 0x%x", opened.Hash, envelope.Hash())
	}
}
func TestEnvelopeAnonymousOpenUntargeted(t *testing.T) {
	payload := []byte("hello envelope")
	envelope, err := NewMessage(payload).Wrap(DefaultPoW, Options{})
	if err != nil {
		t.Fatalf("failed to wrap message: %v", err)
	}
	opened, err := envelope.Open(nil)
	if err != nil {
		t.Fatalf("failed to open envelope: %v", err)
	}
	if opened.To != nil {
		t.Fatalf("recipient mismatch: have 0x%x, want nil", opened.To)
	}
	if !bytes.Equal(opened.Payload, payload) {
		t.Fatalf("payload mismatch: have 0x%x, want 0x%x", opened.Payload, payload)
	}
}
func TestEnvelopeAnonymousOpenTargeted(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}
	payload := []byte("hello envelope")
	envelope, err := NewMessage(payload).Wrap(DefaultPoW, Options{
		To: &key.PublicKey,
	})
	if err != nil {
		t.Fatalf("failed to wrap message: %v", err)
	}
	opened, err := envelope.Open(nil)
	if err != nil {
		t.Fatalf("failed to open envelope: %v", err)
	}
	if opened.To != nil {
		t.Fatalf("recipient mismatch: have 0x%x, want nil", opened.To)
	}
	if bytes.Equal(opened.Payload, payload) {
		t.Fatalf("payload match, should have been encrypted: 0x%x", opened.Payload)
	}
}
func TestEnvelopeIdentifiedOpenUntargeted(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}
	payload := []byte("hello envelope")
	envelope, err := NewMessage(payload).Wrap(DefaultPoW, Options{})
	if err != nil {
		t.Fatalf("failed to wrap message: %v", err)
	}
	opened, err := envelope.Open(key)
	switch err {
	case nil:
		t.Fatalf("envelope opened with bad key: %v", opened)
	case ecies.ErrInvalidPublicKey:
	default:
		t.Fatalf("failed to open envelope: %v", err)
	}
	if opened.To != nil {
		t.Fatalf("recipient mismatch: have 0x%x, want nil", opened.To)
	}
	if !bytes.Equal(opened.Payload, payload) {
		t.Fatalf("payload mismatch: have 0x%x, want 0x%x", opened.Payload, payload)
	}
}
func TestEnvelopeIdentifiedOpenTargeted(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}
	payload := []byte("hello envelope")
	envelope, err := NewMessage(payload).Wrap(DefaultPoW, Options{
		To: &key.PublicKey,
	})
	if err != nil {
		t.Fatalf("failed to wrap message: %v", err)
	}
	opened, err := envelope.Open(key)
	if err != nil {
		t.Fatalf("failed to open envelope: %v", err)
	}
	if opened.To != nil {
		t.Fatalf("recipient mismatch: have 0x%x, want nil", opened.To)
	}
	if !bytes.Equal(opened.Payload, payload) {
		t.Fatalf("payload mismatch: have 0x%x, want 0x%x", opened.Payload, payload)
	}
}

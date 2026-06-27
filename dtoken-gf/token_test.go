package dtoken_gf

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type memoryStore struct {
	data map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{data: make(map[string]string)}
}

func (m *memoryStore) Save(_ context.Context, userKey string, data string) error {
	m.data[userKey] = data
	return nil
}

func (m *memoryStore) Load(_ context.Context, userKey string) (string, error) {
	return m.data[userKey], nil
}

func (m *memoryStore) Delete(_ context.Context, userKey string) error {
	delete(m.data, userKey)
	return nil
}

func TestGenerateReuseTokenSyncsSessionAndResetsLifetime(t *testing.T) {
	store := newMemoryStore()
	token := &DTokenV2{
		Options: Options{MultiLogin: true},
		TokenCodec: &DefaultTokenCodec{
			Delimiter:  DefaultTokenDelimiter,
			EncryptKey: []byte(DefaultEncryptKey),
		},
		SessionCodec: NewDefaultSessionCodec(),
		Store:        store,
	}

	first, err := token.Generate(context.Background(), "user-1", g.Map{"name": "old"})
	if err != nil {
		t.Fatalf("first generate failed: %v", err)
	}

	time.Sleep(time.Millisecond)

	second, err := token.Generate(context.Background(), "user-1", g.Map{"name": "new"})
	if err != nil {
		t.Fatalf("second generate failed: %v", err)
	}
	if second != first {
		t.Fatalf("expected token reuse, got %q and %q", first, second)
	}

	session, err := token.GetSession(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}
	if session.Data["name"] != "new" {
		t.Fatalf("expected session data to be updated, got %#v", session.Data)
	}
	if session.RefreshNum != 0 {
		t.Fatalf("expected refresh num reset, got %d", session.RefreshNum)
	}
	if session.LastRenewTime != 0 {
		t.Fatalf("expected last renew time reset, got %d", session.LastRenewTime)
	}
	if session.CreateTime == 0 {
		t.Fatalf("expected create time to be set")
	}
}

func TestGenerateReuseTokenFallsBackOnMissingSession(t *testing.T) {
	store := newMemoryStore()
	token := &DTokenV2{
		Options: Options{MultiLogin: true},
		TokenCodec: &DefaultTokenCodec{
			Delimiter:  DefaultTokenDelimiter,
			EncryptKey: []byte(DefaultEncryptKey),
		},
		SessionCodec: NewDefaultSessionCodec(),
		Store:        store,
	}

	first, err := token.Generate(context.Background(), "user-2", g.Map{"name": "one"})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if first == "" {
		t.Fatal("expected token")
	}

	if err := token.Destroy(context.Background(), "user-2"); err != nil {
		t.Fatalf("destroy failed: %v", err)
	}

	second, err := token.Generate(context.Background(), "user-2", g.Map{"name": "two"})
	if err != nil {
		t.Fatalf("regenerate failed: %v", err)
	}
	if second == "" {
		t.Fatal("expected token")
	}
	if second == first {
		t.Fatal("expected a new token after destroy")
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	store := newMemoryStore()
	if err := store.Save(context.Background(), "k", "v"); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, ok := store.data["k"]; !ok {
		t.Fatal("expected value")
	}
	if err := store.Delete(context.Background(), "k"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, ok := store.data["k"]; ok {
		t.Fatal("expected value removed")
	}
}

var _ Store = (*memoryStore)(nil)

func TestMemoryStoreEmptyLoad(t *testing.T) {
	store := newMemoryStore()
	got, err := store.Load(context.Background(), "missing")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestNormalizeOptionsAppliesMinTimeout(t *testing.T) {
	opt := NormalizeOptions(Options{Timeout: 5 * 1000})
	if opt.Timeout != MinTimeout {
		t.Fatalf("expected timeout %d, got %d", MinTimeout, opt.Timeout)
	}
	if opt.MaxRefresh <= 0 || opt.MaxRefresh >= opt.Timeout {
		t.Fatalf("expected valid max refresh, got %d with timeout %d", opt.MaxRefresh, opt.Timeout)
	}
}

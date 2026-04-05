package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

// newTestCache spins up an in-memory Redis server and returns a Cache backed by it.
func newTestCache(t *testing.T) (Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return &cache{instance: client, keyPrefix: "test"}, mr
}

var ctx = context.Background()

// ---- Set / Get ----

func TestSet_Get_String(t *testing.T) {
	c, _ := newTestCache(t)
	if err := c.Set(ctx, "name", "alice", WithTTL(time.Minute)); err != nil {
		t.Fatalf("Set: %v", err)
	}
	var got string
	if err := c.Get(ctx, "name", &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "alice" {
		t.Errorf("got %q, want \"alice\"", got)
	}
}

func TestSet_Get_Struct(t *testing.T) {
	type user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "user", user{ID: 1, Name: "alice"})

	var got user
	if err := c.Get(ctx, "user", &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != 1 || got.Name != "alice" {
		t.Errorf("got %+v", got)
	}
}

func TestGet_ErrKeyNotFound(t *testing.T) {
	c, _ := newTestCache(t)
	var s string
	err := c.Get(ctx, "missing", &s)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestSet_NX_ErrKeyNotSet(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "k", "existing")
	err := c.Set(ctx, "k", "new", WithNX())
	if !errors.Is(err, ErrKeyNotSet) {
		t.Errorf("expected ErrKeyNotSet, got %v", err)
	}
}

func TestSet_NX_XX_Conflict(t *testing.T) {
	c, _ := newTestCache(t)
	err := c.Set(ctx, "k", "v", WithNX(), WithXX())
	if err == nil {
		t.Error("expected error for NX+XX combination")
	}
}

func TestGet_DeleteAfterGet(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "token", "abc123")

	var got string
	_ = c.Get(ctx, "token", &got, WithDeleteAfterGet())

	var after string
	err := c.Get(ctx, "token", &after)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("key should be deleted after GetDel, got %v", err)
	}
}

// ---- GetOrSet ----

func TestGetOrSet_CacheMiss(t *testing.T) {
	c, _ := newTestCache(t)
	calls := 0
	var got string
	err := c.GetOrSet(ctx, "computed", &got, func() (any, error) {
		calls++
		return "result", nil
	}, WithTTL(time.Minute))
	if err != nil {
		t.Fatalf("GetOrSet: %v", err)
	}
	if got != "result" {
		t.Errorf("got %q, want \"result\"", got)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestGetOrSet_CacheHit(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "computed", "cached")
	calls := 0
	var got string
	_ = c.GetOrSet(ctx, "computed", &got, func() (any, error) {
		calls++
		return "fresh", nil
	})
	if calls != 0 {
		t.Error("fn should not be called on cache hit")
	}
	if got != "cached" {
		t.Errorf("got %q, want \"cached\"", got)
	}
}

// ---- Delete / Exists ----

func TestDelete(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "a", "1")
	_ = c.Set(ctx, "b", "2")
	if err := c.Delete(ctx, "a", "b"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, _ := c.Exists(ctx, "a")
	if ok {
		t.Error("key should be deleted")
	}
}

func TestExists(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.Set(ctx, "present", "yes")

	ok, err := c.Exists(ctx, "present")
	if err != nil || !ok {
		t.Error("expected key to exist")
	}
	ok, err = c.Exists(ctx, "absent")
	if err != nil || ok {
		t.Error("expected key to not exist")
	}
}

// ---- Increment ----

func TestIncrement(t *testing.T) {
	c, _ := newTestCache(t)
	n, err := c.Increment(ctx, "counter", 1)
	if err != nil || n != 1 {
		t.Errorf("Increment: got %d, %v", n, err)
	}
	n, _ = c.Increment(ctx, "counter", 4)
	if n != 5 {
		t.Errorf("Increment: got %d, want 5", n)
	}
	n, _ = c.Increment(ctx, "counter", -2)
	if n != 3 {
		t.Errorf("Decrement: got %d, want 3", n)
	}
}

// ---- Expire / TTL ----

func TestExpire_TTL(t *testing.T) {
	c, mr := newTestCache(t)
	_ = c.Set(ctx, "k", "v", WithPersist())

	if err := c.Expire(ctx, "k", time.Minute); err != nil {
		t.Fatalf("Expire: %v", err)
	}

	ttl, err := c.TTL(ctx, "k")
	if err != nil {
		t.Fatalf("TTL: %v", err)
	}
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %v", ttl)
	}

	// Fast-forward miniredis clock and check key expires
	mr.FastForward(2 * time.Minute)
	ok, _ := c.Exists(ctx, "k")
	if ok {
		t.Error("key should have expired")
	}
}

// ---- Set operations ----

func TestAddToSet_GetSetMembers(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.AddToSet(ctx, "roles", []string{"admin", "editor"})
	members, err := c.GetSetMembers(ctx, "roles")
	if err != nil {
		t.Fatalf("GetSetMembers: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("got %d members, want 2", len(members))
	}
}

func TestRemoveFromSet_DeletesWhenEmpty(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.AddToSet(ctx, "tags", []string{"go"})
	_ = c.RemoveFromSet(ctx, "tags", []string{"go"})
	members, _ := c.GetSetMembers(ctx, "tags")
	if len(members) != 0 {
		t.Errorf("expected empty set, got %v", members)
	}
}

func TestExistsInSet(t *testing.T) {
	c, _ := newTestCache(t)
	_ = c.AddToSet(ctx, "perms", []string{"read", "write"})

	ok, _ := c.ExistsInSet(ctx, "perms", "read")
	if !ok {
		t.Error("expected 'read' to exist in set")
	}
	ok, _ = c.ExistsInSet(ctx, "perms", "delete")
	if ok {
		t.Error("expected 'delete' to not exist in set")
	}
}

// ---- key prefix ----

func TestKeyPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	c := &cache{instance: client, keyPrefix: "svc"}

	_ = c.Set(ctx, "session", "xyz")
	if !mr.Exists("svc:session") {
		t.Error("expected key to be stored with prefix svc:session")
	}
}

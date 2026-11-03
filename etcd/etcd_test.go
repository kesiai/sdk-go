package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"os"
	"strings"
	"testing"
	"time"
)

// Set SDK_GO_TEST_ETCD_ENDPOINTS to run against a disposable etcd instance.
func testClient(t *testing.T) (*clientv3.Client, context.Context, string) {
	t.Helper()
	endpoints := os.Getenv("SDK_GO_TEST_ETCD_ENDPOINTS")
	if endpoints == "" {
		t.Skip("set SDK_GO_TEST_ETCD_ENDPOINTS to run etcd integration tests")
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		Username:    os.Getenv("SDK_GO_TEST_ETCD_USERNAME"),
		Password:    os.Getenv("SDK_GO_TEST_ETCD_PASSWORD"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	key := fmt.Sprintf("/sdk-go-test/%s/%d", t.Name(), time.Now().UnixNano())
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := client.Delete(cleanupCtx, key); err != nil {
			t.Errorf("cleanup key: %v", err)
		}
	})
	return client, ctx, key
}

func Test_req(t *testing.T) {
	client, ctx, key := testClient(t)
	if _, err := client.Put(ctx, key, "value"); err != nil {
		t.Fatal(err)
	}
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Kvs) != 1 || string(got.Kvs[0].Value) != "value" {
		t.Fatalf("unexpected value: %v", got.Kvs)
	}
	deleted, err := client.Delete(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.Deleted != 1 {
		t.Fatalf("deleted %d keys, want 1", deleted.Deleted)
	}
}

func Test_watch(t *testing.T) {
	client, ctx, key := testClient(t)
	initial, err := client.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	watch := client.Watch(ctx, key, clientv3.WithRev(initial.Header.Revision+1))
	if _, err := client.Put(ctx, key, "watched"); err != nil {
		t.Fatal(err)
	}
	select {
	case response, ok := <-watch:
		if !ok {
			t.Fatal("watch closed before receiving an event")
		}
		if err := response.Err(); err != nil {
			t.Fatal(err)
		}
		if len(response.Events) != 1 || string(response.Events[0].Kv.Value) != "watched" {
			t.Fatalf("unexpected events: %v", response.Events)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func Test_req1(t *testing.T) {
	client, ctx, key := testClient(t)
	lease, err := client.Grant(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := client.Revoke(cleanupCtx, lease.ID); err != nil {
			t.Errorf("revoke lease: %v", err)
		}
	})
	if _, err := client.Put(ctx, key, "leased", clientv3.WithLease(lease.ID)); err != nil {
		t.Fatal(err)
	}
	alive, err := client.KeepAliveOnce(ctx, lease.ID)
	if err != nil {
		t.Fatal(err)
	}
	if alive.ID != lease.ID || alive.TTL <= 0 {
		t.Fatalf("unexpected lease response: %v", alive)
	}
}

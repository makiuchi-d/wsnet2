package lobby

import (
	"testing"
	"time"

	"wsnet2/testdb"
)

func TestGameCache(t *testing.T) {
	lobbyDB := testdb.New("test_game_cache")
	lobbyDB.MustExec(testdb.GetCreateSql("game_server", "../sql/10-schema.sql"))

	now := time.Now()
	nowUnix := now.Unix()
	lobbyDB.MustExec(
		`INSERT INTO game_server (id, hostname, public_name, grpc_port, ws_port, status, heartbeat) VALUES
		(1, "host1", "global1", 1001, 1002, 0, ?),
		(2, "host2", "global2", 2001, 2002, 1, ?),
		(3, "host3", "global3", 3001, 3002, 2, ?),
		(4, "host4", "global4", 4001, 4002, 1, ?)`,
		nowUnix, nowUnix, nowUnix, nowUnix-100)
	// host1 - not ready
	// host2 - ready
	// host3 - shutting down
	// host4 - expired
	// randではhost2のみが選択される
	// Getではhost3も取得可能

	hc := newGameCache(lobbyDB, time.Second, time.Second*10)
	err := hc.update()
	if err != nil {
		t.Fatal(err)
	}
	if hc.lastUpdated.Before(now) {
		t.Errorf("lastUpdated is not updated: now=%v lastUpdated=%v", now, hc.lastUpdated)
	}
	if len(hc.servers) != 2 {
		t.Errorf("len(servers) is not 2: %v", hc.servers)
	}
	if len(hc.order) != 1 {
		t.Errorf("len(order) is not 1: %v", hc.order)
	}
	host, err := hc.Rand()
	if err != nil {
		t.Fatalf("hc.Rand(): %v", err)
	}
	if host == nil {
		t.Fatalf("host is nil")
	}
	if host.Id != 2 {
		t.Errorf("host.Id is not 2: %v", host.Id)
	}
	host2, err := hc.Get(host.Id)
	if err != nil {
		t.Fatalf("hc.Get(%v): %v", host.Id, err)
	}
	if host != host2 {
		t.Errorf("host != host2: %+v != %+v", host, host2)
	}

	host3, err := hc.Get(3)
	if err != nil {
		t.Fatalf("hc.Get(3): %v", err)
	}
	if host3 == nil {
		t.Fatalf("host3 is nil")
	}
}

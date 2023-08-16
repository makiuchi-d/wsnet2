package game

import (
	"context"
	"regexp"
	"testing"
	"time"

	"wsnet2/config"
	"wsnet2/pb"
	"wsnet2/testdb"
)

func TestQueries(t *testing.T) {
	ok, err := regexp.MatchString(
		`INSERT INTO room \((.+,|)id(,.+|)\) VALUES \((.+,|):id(,.+|)\)`,
		roomInsertQuery)
	if err != nil {
		t.Fatalf("roomInsertQuery match error: %+v", err)
	}
	if !ok {
		t.Fatalf("roomInsertQuery not match: %v, %v", ok, roomInsertQuery)
	}

	ok, err = regexp.MatchString(
		`UPDATE room SET (.+,|)app_id=:app_id(,.+|) WHERE id=:id`,
		roomUpdateQuery)
	if err != nil {
		t.Fatalf("roomUpdateQuery match error: %+v", err)
	}
	if !ok {
		t.Fatalf("roomUpdateQuery not match: %v, %v", ok, roomUpdateQuery)
	}
}

func TestNewRoomInfo(t *testing.T) {
	ctx := context.Background()
	db := testdb.New("test_new_roominfo")
	db.MustExec(testdb.ExtractCreateSql("room", "../sql/10-schema.sql"))
	retryCount := 3
	maxNumber := 999

	repo := &Repository{
		app: &pb.App{Id: "testing"},
		conf: &config.GameConf{
			RetryCount: retryCount,
			MaxRoomNum: maxNumber,
		},
		db: db,
	}

	op := &pb.RoomOption{
		Visible:        true,
		Watchable:      false,
		WithNumber:     true,
		SearchGroup:    5,
		ClientDeadline: 30,
		MaxPlayers:     10,
		PublicProps:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
		PrivateProps:   []byte{11, 12, 13, 14, 15, 16, 17, 18},
	}

	// 生成されるはずの値
	seed := time.Now().Unix()
	randsrc.Seed(seed)
	id1 := RandomHex(lenId)
	num1 := randsrc.Int31n(int32(maxNumber)) + 1
	id2 := RandomHex(lenId)
	num2 := randsrc.Int31n(int32(maxNumber)) + 1
	id3 := RandomHex(lenId)
	num3 := randsrc.Int31n(int32(maxNumber)) + 1
	id4 := RandomHex(lenId)
	num4 := randsrc.Int31n(int32(maxNumber)) + 1

	failed := true
	defer func() {
		if failed {
			t.Logf("rooms:\n"+
				"room1: %v, %v\n"+
				"room2: %v, %v\n"+
				"room3: %v, %v\n"+
				"room4: %v, %v\n",
				id1, num1, id2, num2, id3, num3, id4, num4)
		}
	}()

	// 1回目, 3回目のid/numをinsertしておく
	// seedリセット後のnewRoomInfoは2回目のid/numで作られる
	// もう一度リセット後は3回連続失敗する
	insertQuery := "INSERT INTO room" +
		" (id, app_id, host_id, visible, joinable, watchable, number, search_group, max_players, players, watchers, created)" +
		" values (?, 'testing', 1, false, false, false, ?, 1, 1, 1, 0, now())"
	db.MustExec(insertQuery, id1, num1)
	db.MustExec(insertQuery, id3, num3)

	randsrc.Seed(seed)
	tx, _ := db.Beginx()
	ri, err := repo.newRoomInfo(ctx, tx, op)
	if err != nil {
		t.Fatalf("NewRoomInfo fail: %+v", err)
	}

	if ri.Id != id2 {
		t.Fatalf("unexpected Id: %v, wants %v", ri.Id, id2)
	}
	if ri.Number.Number != num2 {
		t.Fatalf("unexpected Number: %v, wants %v", ri.Number.Number, num2)
	}

	randsrc.Seed(seed)
	ri, err = repo.newRoomInfo(ctx, tx, op)
	if err == nil {
		t.Fatalf("NewRoomInfo must be error: id=%v num=%v", ri.Id, ri.Number.Number)
	}

	failed = false
}

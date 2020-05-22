package lobby

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"wsnet2/cachedquery"
	"wsnet2/log"
	"wsnet2/pb"
)

const (
	roomCountQueryString = "SELECT COUNT(*) FROM `room` INNER JOIN `host` USING (`host_id`)"
)

type Room struct {
	*pb.RoomInfo
	Host string `json:"host"`
	URL  string `json:"url"`
}

type RoomService struct {
	db             *sqlx.DB
	grpcPort       int
	wsPort         int
	maxRooms       int
	appQuery       *cachedquery.Query
	roomCountQuery *cachedquery.Query
}

func NewRoom(info *pb.RoomInfo) *Room {
	return &Room{
		RoomInfo: info,
	}
}

func NewRoomService(db *sqlx.DB, grpcPort, wsPort, maxRooms int) *RoomService {
	rs := &RoomService{
		db:       db,
		grpcPort: grpcPort,
		wsPort:   wsPort,
		maxRooms: maxRooms,
	}
	rs.appQuery = cachedquery.New(db, time.Second*5, scanApp, appQueryString)
	rs.roomCountQuery = cachedquery.New(db, time.Millisecond, scanRoomCount, roomCountQueryString)
	return rs
}

func (rs *RoomService) getTotalRooms() (int, error) {
	ir, err := rs.roomCountQuery.Query()
	if err != nil {
		return 0, err
	}
	return ir.(int), nil
}

func scanRoomCount(rows *sqlx.Rows) (interface{}, error) {
	if !rows.Next() {
		panic("failed to fetch room count")
	}
	var roomCount int
	err := rows.Scan(&roomCount)
	return roomCount, err
}

func makeRoomURL(host, roomID string, port int) string {
	return fmt.Sprintf("ws://%s:%d/room/%s", host, port, roomID)
}

func makeRoomHost(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

type gameServer struct {
	id         int
	hostName   string
	publicHost string
}

func (rs *RoomService) Create(appID string, roomOption *pb.RoomOption, clientInfo *pb.ClientInfo) (*Room, error) {
	apps, err := rs.appQuery.Query()
	if err != nil {
		return nil, err
	}
	appExists := false
	for _, app := range apps.([]pb.App) {
		if app.Id == appID {
			appExists = true
			break
		}
	}
	if !appExists {
		return nil, xerrors.Errorf("Unknown appID: %v", appID)
	}

	/* まだhostテーブルが無くて動かない
	nRooms, err := rs.getTotalRooms()
	if err != nil {
		return nil, xerrors.Errorf("failed to fetch room count: %w", err)
	}
	if nRooms >= rs.maxRooms {
		return nil, xerrors.Errorf("maximum number of rooms has been exceeded")
	}
	*/

	// TODO: select game server

	gs := &gameServer{
		id:         1,
		hostName:   "localhost",
		publicHost: "localhost",
	}
	grpcAddr := fmt.Sprintf("%s:%d", gs.hostName, rs.grpcPort)

	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("client connection error: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewGameClient(conn)

	req := &pb.CreateRoomReq{
		AppId:      appID,
		RoomOption: roomOption,
		MasterInfo: clientInfo,
	}

	res, err := client.Create(context.TODO(), req)
	if err != nil {
		fmt.Printf("create room error: %v", err)
		return nil, err
	}

	// TODO: check response

	log.Infof("Created room: %v", res.RoomInfo)

	room := &Room{
		RoomInfo: res.RoomInfo,
	}
	room.Host = makeRoomHost(gs.publicHost, rs.wsPort)
	room.URL = makeRoomURL(gs.publicHost, room.RoomInfo.Id, rs.wsPort)

	return room, nil
}

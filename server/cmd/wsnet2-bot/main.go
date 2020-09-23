package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v4"

	"wsnet2/auth"
	"wsnet2/binary"
	"wsnet2/lobby"
	"wsnet2/lobby/service"
	"wsnet2/pb"
)

var (
	appID  = "testapp"
	appKey = "testapppkey"
)

var (
	WSNet2Version string = "LOCAL"
	WSNet2Commit  string = "LOCAL"
)

type bot struct {
	appId  string
	appKey string
	userId string
	props  binary.Dict
	ws     *websocket.Dialer
}

func NewBot(appId, appKey, userId string) *bot {
	props := binary.Dict{}
	props["p1"] = binary.MarshalInt(65535)
	props["p2"] = binary.MarshalStr16("fuga")

	return &bot{
		appId:  appId,
		appKey: appKey,
		userId: userId,
		props:  props,
		ws: &websocket.Dialer{
			Subprotocols:    []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (b *bot) CreateRoom() (*pb.JoinedRoomRes, error) {
	props := binary.Dict{}
	props["key1"] = binary.MarshalInt(1024)
	props["key2"] = binary.MarshalStr16("hoge")

	param := &service.CreateParam{
		RoomOption: pb.RoomOption{
			Visible:     true,
			Joinable:    true,
			Watchable:   false,
			WithNumber:  true,
			MaxPlayers:  6,
			SearchGroup: 1,
			PublicProps: binary.MarshalDict(props),
		},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	room := &pb.JoinedRoomRes{}

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms", param, room)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[bot:%v] Create success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoom(roomId string, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	room := &pb.JoinedRoomRes{}

	url := fmt.Sprintf("http://localhost:8080/rooms/join/id/%s", roomId)
	err := b.doLobbyRequest("POST", url, param, room)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[bot:%v] Join success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomByNumber(roomNumber int32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	room := &pb.JoinedRoomRes{}

	url := fmt.Sprintf("http://localhost:8080/rooms/join/number/%d", roomNumber)
	err := b.doLobbyRequest("POST", url, param, room)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[bot:%v] Join by room number success, WebSocket=%s\n", b.userId, room.Url)

	return room, nil
}

func (b *bot) JoinRoomAtRandom(searchGroup uint32, queries []lobby.PropQuery) (*pb.JoinedRoomRes, error) {
	param := &service.JoinParam{
		Queries: []lobby.PropQueries{queries},
		ClientInfo: pb.ClientInfo{
			Id:    b.userId,
			Props: binary.MarshalDict(b.props),
		},
	}

	room := &pb.JoinedRoomRes{}

	url := fmt.Sprintf("http://localhost:8080/rooms/join/random/%d", searchGroup)
	err := b.doLobbyRequest("POST", url, param, room)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[bot:%v] Join at random success, WebSocket=%s\n", b.userId, room.Url)
	return room, nil
}

func (b *bot) SearchRoom(searchGroup uint32, queries []lobby.PropQuery) ([]pb.RoomInfo, error) {
	param := &service.SearchParam{
		SearchGroup: searchGroup,
		Queries:     []lobby.PropQueries{queries},
	}

	rooms := []pb.RoomInfo{}

	err := b.doLobbyRequest("POST", "http://localhost:8080/rooms/search", param, &rooms)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[bot:%v] Search success, rooms=%v\n", b.userId, rooms)
	return rooms, nil
}

func (b *bot) doLobbyRequest(method, url string, param, dst interface{}) error {
	var p bytes.Buffer
	err := msgpack.NewEncoder(&p).UseJSONTag(true).Encode(param)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, &p)
	req.Header.Add("Content-Type", "application/x-msgpack")
	req.Header.Add("Host", "localhost")
	req.Header.Add("X-Wsnet-App", b.appId)
	req.Header.Add("X-Wsnet-User", b.userId)

	authdata, err := auth.GenerateAuthData(b.appKey, b.userId, time.Now())
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+authdata)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to lobby request: lobby server returned status %v", res.StatusCode)
	}

	err = msgpack.NewDecoder(res.Body).UseJSONTag(true).Decode(dst)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) DialGame(url, authKey string, seq int) (*websocket.Conn, error) {
	hdr := http.Header{}
	hdr.Add("X-Wsnet-App", b.appId)
	hdr.Add("X-Wsnet-User", b.userId)
	hdr.Add("X-Wsnet-LastEventSeq", strconv.Itoa(seq))

	authdata, err := auth.GenerateAuthData(authKey, b.userId, time.Now())
	if err != nil {
		fmt.Printf("[bot:%v] generate authdata error: %v\n", b.userId, err)
		return nil, err
	}
	hdr.Add("Authorization", "Bearer "+authdata)

	ws, res, err := b.ws.Dial(url, hdr)
	if err != nil {
		fmt.Printf("[bot:%v] dial error: %v, %v\n", b.userId, res, err)
		return nil, err
	}
	fmt.Printf("[bot:%v] response: %v\n", b.userId, res)

	return ws, nil
}

func main() {
	fmt.Println("WSNet2-Bot")
	fmt.Println("WSNet2Version:", WSNet2Version)
	fmt.Println("WSNet2Commit:", WSNet2Commit)

	bot := NewBot(appID, appKey, "12345")

	room, err := bot.CreateRoom()
	if err != nil {
		fmt.Printf("create room error: %v\n", err)
		return
	}

	var queries []lobby.PropQuery
	fmt.Println("key1 =")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 !")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpNot, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 <")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 <=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpLessThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 >")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThan, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)
	fmt.Println("key1 >=")
	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpGreaterThanOrEqual, Val: binary.MarshalInt(1024)}}
	bot.SearchRoom(1, queries)

	ws, err := bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, bot.userId, done)

	queries = []lobby.PropQuery{{Key: "key1", Op: lobby.OpEqual, Val: binary.MarshalInt(1024)}}

	go spawnPlayer(room.RoomInfo.Id, "23456", nil)
	go spawnPlayer(room.RoomInfo.Id, "34567", queries)
	go spawnPlayerByNumber(room.RoomInfo.Number, "45678", nil)
	go spawnPlayerByNumber(room.RoomInfo.Number, "56789", queries)
	go spawnPlayerAtRandom("67890", 1, queries)

	go func() {
		time.Sleep(time.Second * 1)
		spawnPlayer(room.RoomInfo.Id, "78901", nil)
	}()
	go func() {
		time.Sleep(time.Second * 1)
		spawnPlayer(room.RoomInfo.Id, "89012", nil)
	}()

	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("msg 001")
		payload := []byte{byte(binary.MsgTypeSwitchMaster), 0, 0, 1}
		payload = append(payload, binary.MarshalStr8("23456")...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
		time.Sleep(time.Second)
		fmt.Println("msg 002")
		payload = []byte{byte(binary.MsgTypeSwitchMaster), 0, 0, 2}
		payload = append(payload, binary.MarshalStr8("34567")...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
		time.Sleep(time.Second)
		fmt.Println("msg 003")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 1, 2, 3, 4, 5})
		time.Sleep(time.Second)
		fmt.Println("msg 004")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 4, 11, 12, 13, 14, 15})
		time.Sleep(time.Second)
		fmt.Println("msg 005")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 5, 21, 22, 23, 24, 25})
		//time.Sleep(time.Second)
		//fmt.Println("msg 003")
		//ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 3, 21, 22, 23, 24, 25})
		//		time.Sleep(time.Second)
		//		ws.Close()
		time.Sleep(time.Second)
		fmt.Println("msg 006")
		payload = []byte{byte(binary.MsgTypeClientProp), 0, 0, 6}
		payload = append(payload, binary.MarshalDict(binary.Dict{
			"p1": binary.MarshalUShort(20),
			"p2": []byte{},
		})...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
	}()

	//	<-done
	time.Sleep(8 * time.Second)

	fmt.Println("reconnect test")
	ws, err = bot.DialGame(room.Url, room.AuthKey, 2)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done = make(chan bool)
	go eventloop(ws, bot.userId, done)

	go spawnPlayer(room.RoomInfo.Id, "99999", nil)
	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("msg 007")
		payload := []byte{byte(binary.MsgTypeKick), 0, 0, 7}
		payload = append(payload, binary.MarshalStr8("99999")...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
		time.Sleep(time.Second)
		fmt.Println("msg 008")
		payload = []byte{byte(binary.MsgTypeKick), 0, 0, 8}
		payload = append(payload, binary.MarshalStr8("00000")...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
		time.Sleep(time.Second)
		fmt.Println("msg 009")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 9, 31, 32, 33, 34, 35})
		time.Sleep(time.Second)
		fmt.Println("msg 010")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeBroadcast), 0, 0, 10, 41, 42, 43, 44, 45})
		time.Sleep(time.Second)
		fmt.Println("msg 011")
		payload = []byte{byte(binary.MsgTypeSwitchMaster), 0, 0, 11}
		payload = append(payload, binary.MarshalStr8("23456")...)
		ws.WriteMessage(websocket.BinaryMessage, payload)
		time.Sleep(time.Second)
		fmt.Println("msg 012 (leave)")
		ws.WriteMessage(websocket.BinaryMessage, []byte{byte(binary.MsgTypeLeave), 0, 0, 12})
		time.Sleep(time.Second)
		ws.Close()
	}()

	<-done

	time.Sleep(3 * time.Second)
	fmt.Println("reconnect test after leave")
	ws, err = bot.DialGame(room.Url, room.AuthKey, 4)
	if err != nil {
		fmt.Printf("dial game error: %v\n", err)
		return
	}

	done = make(chan bool)
	go eventloop(ws, bot.userId, done)
	<-done
}

func spawnPlayer(roomId, userId string, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId)

	room, err := bot.JoinRoom(roomId, queries)
	if err != nil {
		fmt.Printf("[bot:%v] join room error: %v\n", userId, err)
		return
	}

	ws, err := bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		fmt.Printf("[bot:%v] dial game error: %v\n", userId, err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, userId, done)
}

func spawnPlayerByNumber(roomNumber int32, userId string, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId)

	room, err := bot.JoinRoomByNumber(roomNumber, queries)
	if err != nil {
		fmt.Printf("[bot:%v] join room error: %v\n", userId, err)
		return
	}

	ws, err := bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		fmt.Printf("[bot:%v] dial game error: %v\n", userId, err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, userId, done)
}

func spawnPlayerAtRandom(userId string, searchGroup uint32, queries []lobby.PropQuery) {
	bot := NewBot(appID, appKey, userId)
	room, err := bot.JoinRoomAtRandom(searchGroup, queries)
	if err != nil {
		fmt.Printf("[bot:%v] join room error: %v\n", userId, err)
		return
	}

	ws, err := bot.DialGame(room.Url, room.AuthKey, 0)
	if err != nil {
		fmt.Printf("[bot:%v] dial game error: %v\n", userId, err)
		return
	}

	done := make(chan bool)
	go eventloop(ws, userId, done)
}

func eventloop(ws *websocket.Conn, userId string, done chan bool) {
	defer close(done)
	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("[bot:%v] ReadMessage error: %v\n", userId, err)
			return
		}

		switch ty := binary.EvType(b[0]); ty {
		case binary.EvTypeJoined:
			seqnum := (int(b[1]) << 24) + (int(b[2]) << 16) + (int(b[3]) << 8) + int(b[4])
			namelen := int(b[6])
			name := string(b[7 : 7+namelen])
			props := b[7+namelen:]
			fmt.Printf("[bot:%v] %s: %v %#v, %v, %v\n", userId, ty, seqnum, name, props, b)
		case binary.EvTypePermissionDenied:
			fmt.Printf("[bot:%v] %s: %v\n", userId, ty, b)
		case binary.EvTypeTargetNotFound:
			list, _, err := binary.UnmarshalAs(b[5:], binary.TypeList)
			if err != nil {
				fmt.Printf("[bot:%v] %s: error: %v\n", userId, ty, err)
				break
			}
			fmt.Printf("[bot:%v] %s: %v %v\n", userId, ty, list, b)
		default:
			fmt.Printf("[bot:%v] ReadMessage: %v, %v\n", userId, ty, b)
		}
	}
}
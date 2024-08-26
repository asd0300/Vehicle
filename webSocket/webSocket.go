package websocket

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境下请做更严格的检查
	},
}

var (
	clients    = make(map[*websocket.Conn]string)          // 用于存储连接和用户角色
	clientsMu  sync.Mutex                                  // 用于保护 clients 的互斥锁
	pairings   = make(map[*websocket.Conn]*websocket.Conn) // 存储 customer 与 owner 的配对关系
	pairingsMu sync.Mutex                                  // 用于保护 pairings 的互斥锁
)

// 確認有無線上owner
func hasOtherOwner(excludeConn *websocket.Conn) bool {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn, role := range clients {
		if conn != excludeConn && role == "owner" {
			return true
		}
	}
	return false
}

// 獲取一個owner連結
func getAvailableOwner(excludeConn *websocket.Conn) *websocket.Conn {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn, role := range clients {
		if conn != excludeConn && role == "owner" {
			return conn
		}
	}
	return nil
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set WebSocket upgrade"})
		return
	}
	defer conn.Close()

	// 生成唯一的 WebSocket ID
	websocketID := uuid.New().String()

	role := c.Query("role")

	clientsMu.Lock()
	clients[conn] = role
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()

		// 如果连接断开，从配对中删除
		pairingsMu.Lock()
		if pairedConn, ok := pairings[conn]; ok {
			pairedConn.WriteMessage(websocket.TextMessage, []byte("對方已斷開連結"))
			delete(pairings, pairedConn)
			delete(pairings, conn)
		}
		pairingsMu.Unlock()
	}()

	welcomeMessage := "Connected successfully. Your WebSocket ID is: " + websocketID
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMessage)); err != nil {
		return
	}

	if role == "owner" {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Owner connected. Waiting for customers.")); err != nil {
			return
		}
	} else {
		ownerConn := getAvailableOwner(conn)
		if ownerConn == nil {
			// 没有可用的 owner
			if err := conn.WriteMessage(websocket.TextMessage, []byte("現在沒有客服人員")); err != nil {
				return
			}
		} else {
			// 配對 owner customer
			pairingsMu.Lock()
			pairings[conn] = ownerConn
			pairings[ownerConn] = conn
			pairingsMu.Unlock()
			// 通知 customer 和 owner 已連結
			if err := conn.WriteMessage(websocket.TextMessage, []byte("You are connected to an owner. Start chatting.")); err != nil {
				return
			}
			if err := ownerConn.WriteMessage(websocket.TextMessage, []byte("A customer has been connected to you. Start chatting.")); err != nil {
				return
			}
		}
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		//轉發消息給配對
		pairingsMu.Lock()
		if pairedConn, ok := pairings[conn]; ok {
			pairedConn.WriteMessage(messageType, message)
		}
		pairingsMu.Unlock()
	}
}

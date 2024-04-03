package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type resourcesType map[string]string
type connectionData struct {
	conn          *net.UDPConn
	addr          *net.UDPAddr
	receivedData  string
	receivedBytes int
}

var signalChan = make(chan os.Signal, 1)

const (
	MaxResourceNameSize int = 64
	MaxResponseSize     int = 1024
)

type server struct {
	resourcesInfo resourcesType
	port          string
}

func (s *server) start() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+s.port)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Starting server at: ", addr)

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}

	clientChan := make(chan *connectionData)

	// Для ожидания выполнения всех запущенных горутин перед выходом
	var wg sync.WaitGroup
	stopServer := false
	// Горутина, обрабатывающиая системные прерывания
	go func() {
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		defer func() {
			close(clientChan)
		}()
		sig := <-signalChan
		fmt.Println(sig)
		stopServer = true
	}()

	wg.Add(1)
	// Горутина для чтения запросов от клиента кладет информацию в канал
	go func() {
		defer wg.Done()

		buffer := make([]byte, 1024)
		for !stopServer {
			bytesRead, clientAddr, err := conn.ReadFromUDP(buffer[0:])
			if err != nil {
				fmt.Println("Error reading data: ", err)
				continue
			}
			clientChan <- &connectionData{
				conn:          conn,
				addr:          clientAddr,
				receivedData:  string(buffer[:bytesRead]),
				receivedBytes: bytesRead,
			}
		}
	}()

	// Цикл, обрабатывающий запросы
	for !stopServer {
		clientConn, ok := <-clientChan
		if !ok {
			if stopServer {
				fmt.Println("Can't get data from chan: closed. Interrupt called")
			} else {
				fmt.Println("Error getting data from chan.")
			}
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleClient(clientConn)
		}()
	}

	wg.Wait()
}

func (s *server) handleClient(clientData *connectionData) {
	var response = ""

	// Проверяем, что имя ресурса < 64 байт
	if clientData.receivedBytes > MaxResourceNameSize {
		response = makeErrorResponse("Received resource name size is more than 64 bytes.")
	} else {
		resourceInfo, ok := s.resourcesInfo[clientData.receivedData]
		// Проверяем наличие имени ресурса
		if ok {
			response = makeSuccessResponse(resourceInfo)
			// Проверяем, что содержимое ресурса и обрамление не превышает 1024 байт
			if len(response) > MaxResponseSize {
				response = makeErrorResponse("Resource info is more than 1024 bytes.")
			}
		} else {
			response = makeErrorResponse("No such key " + clientData.receivedData)
		}
	}

	_, err := clientData.conn.WriteToUDP([]byte(response), clientData.addr)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}

	fmt.Println("Sent resource value:\n", response, "to", clientData.addr)
}

func main() {
	var args []string
	if args = os.Args[1:]; len(args) < 2 {
		fmt.Println("Usage: go run client.go <port> <resource_info_path as json file>")
		return
	}
	fmt.Println("Parsing server resources.")

	resourcesInfo, err := getResourcesInfo(args[0])
	if err != nil {
		log.Fatalln(err.Error())
	}

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	s := server{resourcesInfo: *resourcesInfo, port: args[1]}
	s.start()
	close(signalChan)
	fmt.Println("Stopping server.")
}

// Считывает данные из файла. Ожидается, что на входе будет json-структура, чтобы её распарсить в мапу.
func getResourcesInfo(filepath string) (*resourcesType, error) {
	resourcesInfo := make(resourcesType)
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileData, &resourcesInfo)
	return &resourcesInfo, err
}

func makeErrorResponse(errorExplanation string) string {
	return fmt.Sprintf("-ERROR-\n%s\n-END-", errorExplanation)
}

func makeSuccessResponse(resourceInfo string) string {
	return fmt.Sprintf("-BEGIN-\n%s\n-END-", resourceInfo)
}

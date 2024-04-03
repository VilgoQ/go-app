## Build
* `brew install go`
* `go build server.go`
* `go build client.go`
## Usage
Сервер работает с json-форматом. Пример входного файла:
```json
{
  "resource_name": "resource_info",
  "resource_name2": "resource_info2"
}
```
Для запуска:
```bash
./server resource.json <port>
```
```bash
./client <port> resource_name
```

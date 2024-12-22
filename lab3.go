package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Client представляет подключенного клиента
type Client struct {
	Connection net.Conn // Сетевое соединение клиента
	Name       string   // Имя клиента
}

var (
	clients      = make(map[string]*Client) // Глобальная карта клиентов (по имени)
	clientsMutex sync.Mutex                 // Мьютекс для синхронизации доступа к clients
)

func main() {
	// Проверка корректности аргументов при запуске
	if len(os.Args) < 2 {
		fmt.Println("Запуск: go run main.go [server|client] [address:port]")
		return
	}

	mode := os.Args[1]    // Режим запуска (server или client)
	address := os.Args[2] // Адрес и порт для подключения

	if mode == "server" {
		startServer(address)
	} else if mode == "client" {
		startClient(address)
	} else {
		fmt.Println("Некорректный режим. Используйте 'server' или 'client'.")
	}
}

func startServer(address string) {
	listener, err := net.Listen("tcp", address) // Начинаем слушать указанный адрес
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Сервер запущен на", address)

	// Принимаем подключения
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка установки соединения:", err)
			continue
		}

		go handleConnection(connection) // Обработка каждого подключения в отдельной горутине
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close() // Закрыть соединение при завершении функции

	reader := bufio.NewReader(connection)
	name, err := reader.ReadString('\n') // Считываем ник клиента
	if err != nil {
		fmt.Println("Ошибка чтения имени:", err)
		return
	}

	name = strings.TrimSpace(name) // Убираем лишние пробелы и символы новой строки

	// Проверка, существует ли уже клиент с таким именем
	clientsMutex.Lock()
	if _, exists := clients[name]; exists {
		connection.Write([]byte("Имя уже занято. Попробуйте другое.\n"))
		clientsMutex.Unlock()
		return // Завершение обработки, если имя занято
	}

	// Регистрация клиента
	clients[name] = &Client{Connection: connection, Name: name}
	clientsMutex.Unlock()

	fmt.Println(name, "подключился")
	connection.Write([]byte("Добро пожаловать в чат!\n"))

	// Читаем сообщения от клиента
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(name, "отключился")
			clientsMutex.Lock()
			delete(clients, name) // Удаление клиента из списка
			clientsMutex.Unlock()
			return
		}

		handleMessage(name, strings.TrimSpace(message)) // Обработка сообщения
	}
}

func handleMessage(sender string, message string) {
	if strings.HasPrefix(message, "@") {
		// Личное сообщение
		target := strings.SplitN(message, " ", 2) // Разделение на имя получателя и текст сообщения
		if len(target) < 2 {
			return // Игнорируем, если текст отсутствует
		}

		recipientName := strings.TrimPrefix(target[0], "@")
		msg := target[1]

		// Проверка, существует ли получатель
		clientsMutex.Lock()
		recipient, exists := clients[recipientName]
		clientsMutex.Unlock()

		if exists {
			// Отправление сообщения получателю
			recipient.Connection.Write([]byte(fmt.Sprintf("%s: %s\n", sender, msg)))
		}
	} else {
		// Широковещательное сообщение
		broadcastMessage(sender, message)
	}
}

func broadcastMessage(sender string, message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range clients {
		client.Connection.Write([]byte(fmt.Sprintf("%s: %s\n", sender, message)))
	}
}

func startClient(address string) {
	connection, err := net.Dial("tcp", address) // Подключение к серверу
	if err != nil {
		fmt.Println("Ошибка подключения к серверу:", err)
		return
	}
	defer connection.Close()

	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(connection)

	// Ввод имени
	fmt.Print("Введите Ваше имя: ")
	nick, _ := reader.ReadString('\n')
	nick = strings.TrimSpace(nick)
	connection.Write([]byte(nick + "\n"))

	// Получение ответа от сервера
	response, _ := serverReader.ReadString('\n')
	fmt.Print(response) // Вывод сообщения от сервера
	if strings.Contains(response, "имя уже занято") {
		return // Завершение, если имя занято
	}

	fmt.Println("Успешное подключение, введи свое сообщение ниже.")

	// Горутина для получения сообщений от сервера
	go func() {
		for {
			message, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Отключен от сервера.")
				os.Exit(0)
			}
			fmt.Print(message)
		}
	}()

	// Читаем и отправляем сообщения серверу
	for {
		message, _ := reader.ReadString('\n')
		connection.Write([]byte(strings.TrimSpace(message) + "\n"))
	}
}

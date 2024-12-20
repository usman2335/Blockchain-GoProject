package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	// "regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Transaction struct {
	Hash     string
	FileName string
}

type Block struct {
	BlockHash         string
	Transactions      []Transaction
	PreviousBlockHash string
}

var ledger []Block
var nodes []string
var transactionsBuffer []Transaction
var transactions_done int = 0
var mu sync.Mutex
var blockCreated bool = false

func createBlock(transactions []Transaction, previousBlockHash string) Block {
	hasher := sha256.New()
	for _, t := range transactions {
		hasher.Write([]byte(t.Hash))
	}
	hasher.Write([]byte(previousBlockHash))
	blockHash := hex.EncodeToString(hasher.Sum(nil))

	return Block{
		BlockHash:         blockHash,
		Transactions:      transactions,
		PreviousBlockHash: previousBlockHash,
	}
}

func processFile(fileName string) (*Transaction, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", fileName, err)
	}
	defer file.Close()

	var numbers []int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		num, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("failed to parse number in file %s: %v", fileName, err)
		}
		numbers = append(numbers, num)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", fileName, err)
	}

	// Sort the numbers in ascending order
	sort.Ints(numbers)

	// Generate a unique hash based on the sorted numbers
	hasher := sha256.New()
	for _, num := range numbers {
		hasher.Write([]byte(strconv.Itoa(num)))
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	return &Transaction{
		Hash:     hash,
		FileName: fileName,
	}, nil
}
func addTransaction(transaction Transaction) {
	mu.Lock()
	defer mu.Unlock()

	if blockCreated {
		fmt.Println("Block already created. No more transactions will be added.")
		return
	}

	transactionsBuffer = append(transactionsBuffer, transaction)

	if len(transactionsBuffer) == 3 {
		var previousBlockHash string
		if len(ledger) > 0 {
			previousBlockHash = ledger[len(ledger)-1].BlockHash
		} else {
			previousBlockHash = "" // Genesis block
		}

		newBlock := createBlock(transactionsBuffer, previousBlockHash)
		ledger = append(ledger, newBlock)
		transactionsBuffer = []Transaction{} // Clear the buffer
		blockCreated = true

		fmt.Printf("New block created: %v\n", newBlock)
	}
}

// sending Transaction to all nodes
func sendtransaction(nodeIndex int, t Transaction) {

	obj := Transaction{
		Hash:     t.Hash,
		FileName: t.FileName,
	}

	for index, node := range nodes {
		if index == nodeIndex {
			continue
		} else {

			conn, err := net.Dial("tcp", ":"+node)
			if err != nil {
				fmt.Println("Error connecting to server:", err)
				continue
			}
			encoder := gob.NewEncoder(conn)
			err = encoder.Encode(obj)
			if err != nil {
				fmt.Println("Error sending message:", err)
				continue
			}
			conn.Close()
		}
	}
}

//	func readingFiles(nodeID int) {
//		nodeIndex := nodeID - 1
//		for i := nodeID; i <= 1000; i += 4 {
//			fileName := filepath.Join("random_numbers_files", fmt.Sprintf("%d.txt", i))
//			fmt.Printf("Node %d is processing: %s\n", nodeID, fileName)
//			processFile(fileName)
//			transaction, err := processFile(fileName)
//			if err != nil {
//				fmt.Printf("Error processing %s: %v\n", fileName, err)
//			} else {
//				go sendtransaction(nodeIndex, *transaction)
//				fmt.Printf("File: %s, Hash: %s\n", transaction.FileName, transaction.Hash)
//				addTransaction(*transaction) // Add the transaction to the buffer
//				mu.Lock()
//				transactions_done += 1
//				mu.Unlock()
//			}
//		}
//	}
func readingFiles(nodeID int) {
	nodeIndex := nodeID - 1
	for i := nodeID; i <= 1000; i += 4 {
		mu.Lock()
		if blockCreated {
			mu.Unlock()
			break
		}
		mu.Unlock()

		fileName := filepath.Join("random_numbers_files", fmt.Sprintf("%d.txt", i))
		fmt.Printf("Node %d is processing: %s\n", nodeID, fileName)
		transaction, err := processFile(fileName)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", fileName, err)
		} else {
			go sendtransaction(nodeIndex, *transaction)
			fmt.Printf("File: %s, Hash: %s\n", transaction.FileName, transaction.Hash)
			addTransaction(*transaction) // Add the transaction to the buffer
			mu.Lock()
			transactions_done += 1
			mu.Unlock()
		}
	}
}
func incomingTransaction(conn net.Conn) {
	var t Transaction
	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Println("Error decoding message:", err)
		return
	}

	// Add the transaction to the buffer
	addTransaction(t)

	conn.Close()
}

// WILL LISTEN AT 20*1 (* = your node number)
func listenTransactions(nodeIndex int) {
	portnum := nodes[nodeIndex]
	// temp, _ := strconv.Atoi(portnum)
	// temp += 1
	// portnum = strconv.Itoa(temp)

	listener, err := net.Listen("tcp", ":"+portnum)
	if err != nil {
		fmt.Println("Error starting listener on port:", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go incomingTransaction(conn)
	}

}

func nodework(nodeID int, wg *sync.WaitGroup) {

	nodeIndex := nodeID - 1
	go listenTransactions(nodeIndex)
	time.Sleep(30 * time.Second)
	go readingFiles(nodeID)
	for {
		if transactions_done < 1000 {
			continue
		} else {
			break
		}
	}

	wg.Done()

}

func printLedger() {
	mu.Lock() // Lock the mutex to ensure thread safety
	defer mu.Unlock()

	fmt.Println("\n--- Ledger ---")
	for index, hash := range ledger {
		fmt.Printf("File: %d, Hash: %s\n", index, hash)
	}
	fmt.Println("----------------\n")
}

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	if len(os.Args) < 2 {
		fmt.Println("No arguments passed")
		return
	}

	nodes = append(nodes, "2010")
	nodes = append(nodes, "2020")
	nodes = append(nodes, "2030")
	nodes = append(nodes, "2040")

	arg := os.Args[1]
	nodeID, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Println("Error converting argument to integer:", err)
	}

	go nodework(nodeID, &wg)

	go func() {
		for {
			time.Sleep(10 * time.Second) // Print every 10 seconds
			printLedger()
		}
	}()

	wg.Wait()

	time.Sleep(10 * time.Second)
}

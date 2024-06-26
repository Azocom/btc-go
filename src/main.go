package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"btcgo/src/crypto/btc_utils"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

// Define an interface for your handlers
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Implement a simple handler
type HelloHandler struct{}

func (h HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

// Wallets struct to hold the array of wallet addresses
type Wallets struct {
	Addresses [][]byte `json:"wallets"`
}

// Range struct to hold the minimum, maximum, and status
type Range struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Status int    `json:"status"`
}

// Ranges struct to hold an array of ranges
type Ranges struct {
	Ranges []Range `json:"ranges"`
}

var (
	keysChecked   int
	startTime     time.Time
	numCPU        int
	privKeyChan   chan *big.Int
	resultChan    chan *big.Int
	wg            sync.WaitGroup
	stopChan      chan bool
	running       bool
	mu            sync.Mutex
	currentRange  Range
	wallets       *Wallets
	privKeyMinStr string
	privKeyMaxStr string
)

func main() {
	green := color.New(color.FgGreen).SprintFunc()

	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Erro ao obter o caminho do executável: %v\n", err)
		return
	}
	rootDir := filepath.Dir(exePath)

	color.Cyan("BTC GO - Investidor Internacional")
	color.White("v0.123")

	// Load wallet addresses from JSON file
	wallets, err = LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	keysChecked = 0
	startTime = time.Now()

	// Number of CPU cores to use
	numCPU = runtime.NumCPU() 
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU * 2)

	// Initialize channels
	privKeyChan = make(chan *big.Int)
	resultChan = make(chan *big.Int)
	stopChan = make(chan bool)
	running = false

	// Start the web server
	http.HandleFunc("/", webHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/keys", keysHandler) // New route for keys
	log.Println("Starting web server on :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func worker() {
	defer wg.Done()
	for {
		select {
		case privKeyInt, ok := <-privKeyChan:
			if !ok {
				return
			}
			address := btc_utils.CreatePublicHash160(privKeyInt)
			if Contains(wallets.Addresses, address) {
				saveResult(privKeyInt)
				resultChan <- privKeyInt
				return
			}
		case <-stopChan:
			return
		}
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if running {
		http.Error(w, "Process already running", http.StatusBadRequest)
		return
	}

	rangeMin := r.FormValue("rangeMin")
	rangeMax := r.FormValue("rangeMax")
	if rangeMin == "" || rangeMax == "" {
		http.Error(w, "Invalid range values", http.StatusBadRequest)
		return
	}

	// Strip "0x" prefix if present
	if strings.HasPrefix(rangeMin, "0x") {
		rangeMin = rangeMin[2:]
	}
	if strings.HasPrefix(rangeMax, "0x") {
		rangeMax = rangeMax[2:]
	}

	privKeyMin := new(big.Int)
	privKeyMax := new(big.Int)
	_, successMin := privKeyMin.SetString(rangeMin, 16)
	_, successMax := privKeyMax.SetString(rangeMax, 16)
	if !successMin || !successMax {
		http.Error(w, "Invalid range values", http.StatusBadRequest)
		return
	}

	if privKeyMin.Cmp(privKeyMax) >= 0 {
		http.Error(w, "Invalid range values: min should be less than max", http.StatusBadRequest)
		return
	}

	privKeyMinStr = rangeMin
	privKeyMaxStr = rangeMax
	running = true
	go startProcessing(privKeyMin, privKeyMax)

	fmt.Fprintln(w, "Processing started")
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		http.Error(w, "No process running", http.StatusBadRequest)
		return
	}

	running = false
	close(stopChan)
	fmt.Fprintln(w, "Processing stopped")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	status := map[string]interface{}{
		"keysChecked":   keysChecked,
		"elapsedTime":   time.Since(startTime).Seconds(),
		"keysPerSecond": float64(keysChecked) / time.Since(startTime).Seconds(),
		"running":       running,
		"rangeMin":      "0x" + privKeyMinStr,
		"rangeMax":      "0x" + privKeyMaxStr,
	}

	// Permitir solicitações de qualquer origem
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Definir o tipo de conteúdo da resposta como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	 // Permitir solicitações de qualquer origem
	 w.Header().Set("Access-Control-Allow-Origin", "*")
	 w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	 w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	fmt.Fprintf(w, `<html>
		<head><title>BTC Go</title></head>
		<body>
			<h1>BTC Go</h1>
			<form action="/start" method="post">
				<label for="rangeMin">Enter Start Range (hex):</label>
				<input type="text" id="rangeMin" name="rangeMin">
				<br>
				<label for="rangeMax">Enter End Range (hex):</label>
				<input type="text" id="rangeMax" name="rangeMax">
				<br><br>
				<input type="submit" value="Start Processing">
			</form>
			<br>
			<p><a href="/stop">Stop Processing</a></p>
			<p><a href="/status">Check Status</a></p>
			<p><a href="/keys">View Found Keys</a></p> <!-- Link to view found keys -->
		</body>
	</html>`)
}

func startProcessing(privKeyMin, privKeyMax *big.Int) {
	// Reset channels and wait group
	privKeyChan = make(chan *big.Int)
	resultChan = make(chan *big.Int)
	stopChan = make(chan bool)

	// Start worker goroutines
	for i := 0; i < numCPU*3; i++ {
		wg.Add(1)
		go worker()
	}

	// Ticker for periodic updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Goroutine to print speed updates
	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				elapsedTime := time.Since(startTime).Seconds()
				keysPerSecond := float64(keysChecked) / elapsedTime
				fmt.Printf("Chaves checadas: %s, Chaves por segundo: %s\n", humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)))
				mu.Unlock()
			case <-stopChan:
				return
			}
		}
	}()

	// Send private keys to the workers
	go func() {
		for privKeyMin.Cmp(privKeyMax) < 0 {
			select {
			case <-stopChan:
				close(privKeyChan)
				return
			default:
				privKeyCopy := new(big.Int).Set(privKeyMin)
				privKeyChan <- privKeyCopy
				privKeyMin.Add(privKeyMin, big.NewInt(1))
				mu.Lock()
				keysChecked++
				mu.Unlock()
			}
		}
		close(privKeyChan)
	}()

	// Wait for a result from any worker
	var foundAddress *big.Int
	select {
	case foundAddress = <-resultChan:
		color.Yellow("Chave privada encontrada: %064x\n", foundAddress)
		color.Yellow("WIF: %s", btc_utils.GenerateWif(foundAddress))
		saveResult(foundAddress)
		close(stopChan)
	}

	// Wait for all workers to finish
	wg.Wait()

	mu.Lock()
	running = false
	mu.Unlock()
}

func saveResult(privKey *big.Int) {
	type Result struct {
		PrivateKey string `json:"privateKey"`
		WIF        string `json:"wif"`
	}

	wif := btc_utils.GenerateWif(privKey)
	result := Result{
		PrivateKey: fmt.Sprintf("%064x", privKey),
		WIF:        wif,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result to JSON: %v", err)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		return
	}
	rootDir := filepath.Dir(exePath)

	resultFilePath := filepath.Join(rootDir, "data", "result.json")
	file, err := os.Create(resultFilePath)
	if err != nil {
		log.Printf("Error creating result file: %v", err)
		return
	}
	defer file.Close()

	_, err = file.Write(resultJSON)
	if err != nil {
		log.Printf("Error writing result to file: %v", err)
		return
	}

	log.Printf("Result saved to %s", resultFilePath)
}

func keysHandler(w http.ResponseWriter, r *http.Request) {
	exePath, err := os.Executable()
	if err != nil {
		http.Error(w, "Error getting executable path", http.StatusInternalServerError)
		return
	}
	rootDir := filepath.Dir(exePath)

	resultFilePath := filepath.Join(rootDir, "data", "result.json")
	file, err := os.Open(resultFilePath)
	if err != nil {
		http.Error(w, "Error reading result file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode JSON from file
	var result struct {
		PrivateKey string `json:"privateKey"`
		WIF        string `json:"wif"`
	}
	err = json.NewDecoder(file).Decode(&result)
	if err != nil {
		http.Error(w, "Error decoding result file", http.StatusInternalServerError)
		return
	}

	// Encode and send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// LoadRanges and LoadWallets functions should be implemented to load the data from the respective JSON files.

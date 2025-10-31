package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Configuration constants
const (
	// Maximum number of future words associated with a word
	MaxLeaf = 30
	// Network port: binds to 127.0.0.1:PORT
	Port = 9001
	// Maximum data sent in a single HTTP chunk
	BufferSize = 1024 * 5
	// Number of words in one paragraph of text. Periods are counted as words
	WordCount = 200
	// Number of paragraphs
	PCount = 3
)

// Text inserted into 1/4th of pages
const Poison = ""

// Directory to which the babbler will link.
// Must begin and end with /s
const URLPrefix = "/babble/"

// Global statistics
var (
	requestsServed uint64
	bytesServed    uint64
	start          time.Time
)

// MarkovWord represents a single word's entry in the Markov chain
type MarkovWord struct {
	// The actual word
	Key string
	// Child words, as strings and indices into the chain
	Values      []string
	ValuesIndex []int
}

// MarkovChain represents the whole Markov chain
type MarkovChain struct {
	// The index of the sentence separator "END"
	StartKey int
	// All the words
	Keys []MarkovWord
}

// NewChain creates a new MarkovChain
func NewChain() *MarkovChain {
	return &MarkovChain{
		StartKey: -1,
		Keys:     []MarkovWord{},
	}
}

// LoadFile loads a Markov chain from a file
func LoadFile(filename string) *MarkovChain {
	log.Printf("    Loading %s...\n", filename)
	chain := NewChain()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		entry := MarkovWord{
			Key:         words[0],
			Values:      []string{},
			ValuesIndex: []int{},
		}

		// Add all child words (maximum of MaxLeaf)
		for i := 1; i < len(words) && len(entry.Values) < MaxLeaf; i++ {
			if words[i] != "" {
				entry.Values = append(entry.Values, words[i])
			}
		}

		chain.Keys = append(chain.Keys, entry)

		// Save index if we just parsed the sentence separator
		if entry.Key == "END" {
			chain.StartKey = len(chain.Keys) - 1
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Precompute the indices of the next words
	for i := range chain.Keys {
		entry := &chain.Keys[i]
		entry.ValuesIndex = make([]int, len(entry.Values))
		for e, thisWord := range entry.Values {
			found := false
			for linked, word := range chain.Keys {
				if word.Key == thisWord {
					entry.ValuesIndex[e] = linked
					found = true
					break
				}
			}
			// Sanity check: Makes sure a matching entry exists
			if !found {
				log.Fatalf("No matching entry found for word: %s", thisWord)
			}
		}
	}

	// Truncate words at hyphens to allow hacking in high-order markov chains
	for i := range chain.Keys {
		entry := &chain.Keys[i]
		if hyphenIndex := strings.Index(entry.Key, "-"); hyphenIndex != -1 {
			entry.Key = entry.Key[:hyphenIndex]
		}
	}

	// Sanity check: will fail if the sentence separator wasn't in the chain
	if chain.StartKey == -1 {
		log.Fatal("Sentence separator 'END' not found in chain")
	}
	return chain
}

// HashString is a non-secure hash function used to seed the RNG
func HashString(s string) uint32 {
	acc := uint32(0xDEADBEEF)
	for i := 0; i < len(s); i++ {
		acc += uint32(s[i])
		acc *= 13
		acc = acc << 8
		acc %= ((uint32(1) << 31) - 1)
	}
	return acc
}

// PRNG is an XORSHIFT style RNG
func PRNG(state *uint32) int {
	x := *state
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*state = x
	return int(x)
}

// RandomText generates length words using the markov chain, returns the words
func RandomText(chain *MarkovChain, length int, seed *uint32) string {
	buff := bytes.NewBufferString("")
	SendText(chain, length, buff, seed)
	return buff.String()
}

// SendText generates length words using the markov chain, writes results into the buffer
func SendText(chain *MarkovChain, length int, dst io.Writer, seed *uint32) {
	nextIndex := chain.StartKey
	capitalize := true

	for i := 0; i < length; i++ {
		// Pick a next word at random
		r := float64(PRNG(seed)%900) / 900
		r = r * r
		r = r * float64(len(chain.Keys[nextIndex].Values))

		// Bounds check
		selection := int(r)
		if selection < 0 {
			selection = 0
		}
		if selection >= len(chain.Keys[nextIndex].Values) {
			selection = len(chain.Keys[nextIndex].Values) - 1
		}

		// Advance chain
		nextIndex = chain.Keys[nextIndex].ValuesIndex[selection]

		// Check for "END"
		if len(chain.Keys[nextIndex].Key) > 0 && chain.Keys[nextIndex].Key[0] == 'E' {
			if !capitalize {
				dst.Write([]byte("."))
				capitalize = true
			}
		} else {
			word := chain.Keys[nextIndex].Key
			dst.Write([]byte(" "))
			if capitalize {
				capitalize = false
				dst.Write([]byte(strings.ToUpper(word)))
			} else {
				dst.Write([]byte(word))
			}
		}
	}
}

// RandomWord randomly selects a word. Used for links and topics
func RandomWord(chain *MarkovChain, seed *uint32) string {
	index := PRNG(seed) % len(chain.Keys)
	if int(index) == chain.StartKey {
		return "jellyfish"
	}
	return chain.Keys[index].Key
}

func formatTime(buff io.Writer, seconds int64) {
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	years := days / 365

	if years > 0 {
		buff.Write([]byte(fmt.Sprintf("%d years ", years)))
	}
	if days > 0 {
		buff.Write([]byte(fmt.Sprintf("%d days ", days%365)))
	}
	if hours > 0 {
		buff.Write([]byte(fmt.Sprintf("%d hours ", hours%24)))
	}
	if minutes > 0 {
		buff.Write([]byte(fmt.Sprintf("%d minutes ", minutes%60)))
	}
	buff.Write([]byte(fmt.Sprintf("%d seconds ", seconds%60)))
}

func formatNumber(buff io.Writer, number uint64, si bool) {
	if number == 0 {
		buff.Write([]byte("0 "))
		return
	}

	prefix := int(math.Log10(float64(number))) / 3
	n := number
	for i := 0; i < prefix; i++ {
		n /= 1000
	}

	buff.Write([]byte(fmt.Sprintf("%d ", n)))

	if si {
		switch prefix {
		case 1:
			buff.Write([]byte("k"))
		case 2:
			buff.Write([]byte("M"))
		case 3:
			buff.Write([]byte("G"))
		case 4:
			buff.Write([]byte("T"))
		}
	} else {
		switch prefix {
		case 1:
			buff.Write([]byte("thousand "))
		case 2:
			buff.Write([]byte("million "))
		case 3:
			buff.Write([]byte("billion "))
		case 4:
			buff.Write([]byte("trillion "))
		}
	}
}

// Global chains
var allChains []*MarkovChain

func babbleLink(seed uint32) string {
	chain := allChains[PRNG(&seed)%len(allChains)]
	parts := make([]string, 0)
	var i uint32 = 0
	for range 3 {
		i += 1
		seed += i
		parts = append(parts, RandomWord(chain, &seed))
	}
	return fmt.Sprintf("/babble/%s/%s/%s", parts[0], parts[1], parts[2])
}

func babbleTitle(seed uint32) string {
	chain := allChains[PRNG(&seed)%len(allChains)]
	return RandomText(chain, 10, &seed)
}

func babbleLinks(path string) []Link {
	seed := HashString(path)
	result := make([]Link, 0)
	var i uint32 = 0
	for range 5 {
		i += 1
		result = append(result, Link{
			Href:  babbleLink(seed + i),
			Title: babbleTitle(seed + i),
		})
	}
	return result
}
func handleBabbleRequest(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&requestsServed, 1)

	// Extract counter value from the path
	ctr := 0
	for _, c := range r.URL.Path {
		if c >= '0' && c <= '9' {
			ctr, _ = strconv.Atoi(string(c))
			break
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	if len(allChains) == 0 {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("<html><body>Still loading...</body></html>"))
		return
	}
	// Content buffer for chunked encoding
	seed := HashString(r.URL.Path)

	if strings.HasPrefix(r.URL.Path, "/status/") {
		now := time.Now()
		elapsed := now.Sub(start).Seconds()

		w.Write([]byte("<html><head><style>"))
		w.Write([]byte("body {color: white; background-color: black}"))
		w.Write([]byte("div {max-width: 40em; margin: auto;}"))
		w.Write([]byte("h3, h1 {text-align: center}"))
		w.Write([]byte("a {color: cyan;}"))
		w.Write([]byte("</style>"))
		w.Write([]byte("<title>Babbler status</title>"))
		w.Write([]byte("</head><body>"))
		w.Write([]byte("<h1>Babbler stats:</h1>"))
		w.Write([]byte("<div><p>In the past <b>"))
		formatTime(w, int64(elapsed))
		w.Write([]byte("</b>I've spent <b>"))
		formatTime(w, int64(elapsed)) // In Go we don't track CPU time separately
		w.Write([]byte("</b>dealing with: <b>"))
		formatNumber(w, atomic.LoadUint64(&requestsServed), false)
		w.Write([]byte("</b>requests and serving <b>"))
		formatNumber(w, atomic.LoadUint64(&bytesServed), true)
		w.Write([]byte("B</b> of garbage.<br><br>... at an average rate of <b>"))

		if elapsed > 0 {
			perMin := float64(atomic.LoadUint64(&requestsServed)) * 60 / elapsed
			formatNumber(w, uint64(perMin), true)
			w.Write([]byte("</b>requests per minute and <b>"))
			perMin = float64(atomic.LoadUint64(&bytesServed)) * 60 / elapsed
			formatNumber(w, uint64(perMin), true)
			w.Write([]byte("B</b> per minute.<br><br></div>"))
		}
	} else {
		// Pick which chain to use at random
		chain := allChains[PRNG(&seed)%len(allChains)]

		// What do we write about?
		topics := []string{RandomWord(chain, &seed), RandomWord(chain, &seed)}

		// Write HTML
		w.Write([]byte("<html><head><meta http-equiv='Content-Type' content='text/html; charset=UTF-8' /><style>"))
		w.Write([]byte("body {color: white; background-color: black}"))
		w.Write([]byte("div {max-width: 40em; margin: auto;}"))
		w.Write([]byte("h3, h1 {text-align: center}"))
		w.Write([]byte("a {color: cyan;}"))
		w.Write([]byte("</style>"))
		w.Write([]byte("<title>"))
		w.Write([]byte(strings.ToUpper(topics[0])))
		w.Write([]byte(" "))
		w.Write([]byte(strings.ToUpper(topics[1])))
		w.Write([]byte("</title></head><body><h1>"))
		w.Write([]byte(strings.ToUpper(topics[0])))
		w.Write([]byte(" "))
		w.Write([]byte(strings.ToUpper(topics[1])))
		w.Write([]byte("</h1><h3>Garbage for the garbage king!</h3><div>"))

		// Write paragraphs
		for i := 0; i < PCount; i++ {
			w.Write([]byte("<p>"))
			SendText(chain, WordCount, w, &seed)
			w.Write([]byte(".</p>"))
		}

		// Add bonus text if needed
		if PRNG(&seed)%4 == 0 {
			w.Write([]byte("<p>"))
			w.Write([]byte(Poison))
			w.Write([]byte("</p>"))
		}

		// Links
		for i := 0; i < 5; i++ {
			// Link URL
			w.Write([]byte("<a href="))
			w.Write([]byte(URLPrefix))
			w.Write([]byte(RandomWord(chain, &seed)))
			w.Write([]byte("/"))
			w.Write([]byte(RandomWord(chain, &seed)))
			w.Write([]byte("/"))
			w.Write([]byte(RandomWord(chain, &seed)))
			w.Write([]byte("/"))
			w.Write([]byte(RandomWord(chain, &seed)))
			w.Write([]byte("/"))
			w.Write([]byte(RandomWord(chain, &seed)))

			// Embed counter
			ctrString := fmt.Sprintf("/%d/", ctr+1)
			w.Write([]byte(ctrString))

			// Add link text
			w.Write([]byte(" >"))
			SendText(chain, 10, w, &seed)
			w.Write([]byte("</a><br/>"))
		}

		// Footer
		w.Write([]byte("</div></body></html>"))
	}

	// Make sure no data remains in the buffer
	//w.Flush()

	// Send zero length chunk to tell the client that we are done
	fmt.Fprintf(w, "0\r\n\r\n")
}

func loadBabbler() {
	log.Println("[*] Loading files")
	allChains = append(allChains, LoadFile("chain1.txt"))
	allChains = append(allChains, LoadFile("chain2.txt"))
	allChains = append(allChains, LoadFile("chain3.txt"))

	log.Println("[*] Creating server")
	start = time.Now()
}

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entity struct {
	Path      string
	Frequency int
	LastVisit int64 // unix time
}

type Entities []Entity

// version is set at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

// score returns a frecency score combining frequency and recency.
func (e Entity) score() float64 {
	age := time.Now().Unix() - e.LastVisit
	if age < 3600 {
		return float64(e.Frequency) * 4
	} else if age < 86400 {
		return float64(e.Frequency) * 2
	} else if age < 604800 {
		return float64(e.Frequency) * 0.5
	}
	return float64(e.Frequency) * 0.25
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "-h", "--help", "help":
		printHelp()
	case "-v", "--version", "version":
		fmt.Println("jumper " + version)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("usage: jumper add <path>")
			os.Exit(1)
		}
		if err := add(os.Args[2]); err != nil {
			fmt.Println("Error adding path: ", err)
			os.Exit(1)
		}
	case "query":
		if len(os.Args) < 3 {
			fmt.Println("usage: jumper query <term>")
			os.Exit(1)
		}
		path, err := query(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(path)
	case "list":
		if err := list(); err != nil {
			fmt.Println("Error listing history: ", err)
			os.Exit(1)
		}
	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`jumper - frecency-based directory history tracker

Usage:
  jumper add <path>     record a visit to <path>
  jumper query <term>   print the best-matching path for <term>
  jumper list           print all tracked paths, ranked by frecency
  jumper --help         show this help

Note: jumper only tracks and queries history, it does not change your
shell's directory. Use the 'j' shell function (see jumper.sh) to cd.`)
}

// historyPath return the history file path
func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cannot find home directory!")
	}

	return filepath.Join(home, ".local", "share", "jumper", "history.csv")
}

// loadEntities reads the history file into memory. Missing file is not an error.
func loadEntities() (Entities, error) {
	path := historyPath()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Entities{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading history CSV: %w", err)
	}

	entities := make(Entities, 0, len(records))
	for _, row := range records {
		if len(row) != 3 {
			continue
		}
		freq, err := strconv.Atoi(row[1])
		if err != nil {
			continue
		}
		last, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			continue
		}
		entities = append(entities, Entity{Path: row[0], Frequency: freq, LastVisit: last})
	}

	return entities, nil
}

// saveEntities writes entities back to the history file, creating dirs as needed.
// It writes to a temp file in the same directory, fsyncs it, then renames it
// over the real file so a crash or power loss mid-write can't corrupt or
// truncate the existing history.
func saveEntities(entities Entities) error {
	path := historyPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating history dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".history-*.csv.tmp")
	if err != nil {
		return fmt.Errorf("creating temp history file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once the rename below succeeds

	writer := csv.NewWriter(tmp)
	for _, e := range entities {
		row := []string{e.Path, strconv.Itoa(e.Frequency), strconv.FormatInt(e.LastVisit, 10)}
		if err := writer.Write(row); err != nil {
			tmp.Close()
			return fmt.Errorf("writing history row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		tmp.Close()
		return fmt.Errorf("flushing history rows: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp history file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp history file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replacing history file: %w", err)
	}

	return nil
}

// add records a visit to path, bumping frequency and last-visit time.
func add(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	entities, err := loadEntities()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	found := false
	for i := range entities {
		if entities[i].Path == abs {
			entities[i].Frequency++
			entities[i].LastVisit = now
			found = true
			break
		}
	}

	if !found {
		entities = append(entities, Entity{Path: abs, Frequency: 1, LastVisit: now})
	}

	return saveEntities(entities)
}

// query returns the best matching stored path for term, ranked by frecency.
func query(term string) (string, error) {
	entities, err := loadEntities()
	if err != nil {
		return "", err
	}

	term = strings.ToLower(term)
	matches := make(Entities, 0)
	for _, e := range entities {
		if strings.Contains(strings.ToLower(e.Path), term) {
			matches = append(matches, e)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no match found for %q", term)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score() > matches[j].score()
	})

	return matches[0].Path, nil
}

// list prints all stored entities ranked by frecency.
func list() error {
	entities, err := loadEntities()
	if err != nil {
		return err
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].score() > entities[j].score()
	})

	for _, e := range entities {
		fmt.Printf("%-10.2f %s\n", e.score(), e.Path)
	}

	return nil
}

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/internal"
	"gopkg.in/yaml.v3"
)

const Version = "0.4.2"

const usage = `Usage of scaleconnect:

  -c, --config       Path to config file
  -db                Store loaded weights in SQLite
  -db-path           Path to SQLite database file
  -i, --interactive  Keep STDIN open
  -r, --repeat       Run config every N time (format: 2h45m)
`

func main() {
	var (
		config      string
		repeat      string
		interactive bool
		db          bool
		dbPath      string
	)

	flag.Usage = func() { fmt.Print(usage) }
	flag.StringVar(&config, "config", "", "")
	flag.StringVar(&config, "c", "", "")
	flag.StringVar(&repeat, "repeat", "", "")
	flag.StringVar(&repeat, "r", "", "")
	flag.BoolVar(&interactive, "interactive", false, "")
	flag.BoolVar(&interactive, "i", false, "")
	flag.BoolVar(&db, "db", true, "")
	flag.StringVar(&dbPath, "db-path", "", "")
	flag.Parse()

	log.Printf("scaleconnect version %s\n", Version)

	data, err := readConfig(config)
	if err == nil && data != nil {
		if count, err := countSyncs(data); err == nil {
			log.Printf("scaleconnect config: %d sync(s)\n", count)
		}
	}

	requests := make(chan processRequest, 10)
	closers, mqttErr := setupHAMQTTResets(data, requests)
	if mqttErr != nil {
		log.Printf("ha_mqtt reset setup error: %v\n", mqttErr)
	}

	// run config once
	if repeat == "" && !interactive && len(closers) == 0 {
		if err != nil {
			log.Fatal(err)
		}

		if err = process(data, db, dbPath); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	if data != nil {
		requests <- processRequest{Data: data}

		if repeat != "" {
			var sleep time.Duration
			if sleep, err = time.ParseDuration(repeat); err != nil {
				log.Fatal(err)
			}

			go func() {
				for range time.NewTicker(sleep).C {
					requests <- processRequest{Data: data}
				}
			}()
		}
	}

	if interactive {
		go func() {
			// read stdin and process it forever
			reader := bufio.NewReader(os.Stdin)
			for {
				data, err := reader.ReadBytes('\n')
				if err != nil {
					break
				}
				requests <- processRequest{Data: data}
			}
		}()
	}

	go func() {
		for req := range requests {
			if req.Data != nil {
				if err = process(req.Data, db, dbPath); err != nil {
					log.Fatal(err)
				}
				continue
			}

			if err = processSync(req.Name, req.Config, db, dbPath); err != nil {
				log.Fatal(err)
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	for _, close := range closers {
		close()
	}
	fmt.Printf("exit with signal: %s\n", sig)
}

const configName = "scaleconnect.yaml"

func readConfig(name string) ([]byte, error) {
	if name != "" {
		// 1. Check if JSON passed as config
		if name[0] == '{' {
			return []byte(name), nil
		}

		// 2. Check config from passed path
		return os.ReadFile(name)
	}

	// 3. Check config file in CWD
	if data, err := os.ReadFile(configName); err == nil {
		return data, nil
	}

	// 4. Check config near binary
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	path := filepath.Dir(ex)

	data, err := os.ReadFile(filepath.Join(path, configName))
	if err != nil {
		return nil, err
	}

	// change CWD so json file will be near app
	return data, os.Chdir(path)
}

func process(data []byte, db bool, dbPath string) error {
	var syncs map[string]syncConfig
	if err := yaml.Unmarshal(data, &syncs); err != nil {
		return err
	}

	for name, v := range syncs {
		if err := processSync(name, v, db, dbPath); err != nil {
			return err
		}
	}

	return nil
}

func processSync(name string, v syncConfig, db bool, dbPath string) error {
	if v.From == "" || v.To == "" {
		return nil
	}

	syncID, err := getSyncID(name, v)
	if err != nil {
		log.Printf("%s: calc sync id error: %v\n", name, err)
		return nil
	}

	mqttTarget := internal.IsHAMQTT(v.To)
	dbEnabled := db || mqttTarget
	if dbEnabled {
		dbPath, err = normalizeDBPath(dbPath)
		if err != nil {
			return err
		}
	}

	weights, err := internal.GetWeights(v.From)
	if err != nil {
		log.Printf("%s: load data error: %v\n", name, err)
		return nil
	}

	var stats internal.StoreStats
	if dbEnabled {
		stats, err = internal.StoreWeights(dbPath, syncID, weights)
		if err != nil {
			log.Printf("%s: db write error: %v\n", name, err)
			return nil
		}
		log.Printf("%s: sync_id=%s synced=%d new=%d\n", name, syncID, stats.Synced, stats.New)
	} else {
		log.Printf("%s: sync_id=%s synced=%d new=0 db=false\n", name, syncID, len(weights))
	}

	if v.Expr != nil {
		if err = internal.Expr(v.Expr, weights); err != nil {
			log.Printf("%s: calc expr error: %v\n", name, err)
			return nil
		}
	}

	if mqttTarget {
		if stats.New > 0 {
			if err = internal.PublishHAMQTT(v.To, name, syncID, stats.NewWeights); err != nil {
				log.Printf("%s: mqtt publish error: %v\n", name, err)
				return nil
			}
		} else {
			log.Printf("%s: mqtt state skipped because new=0\n", name)
		}
		log.Printf("%s: OK\n", name)
		return nil
	}

	if err = internal.SetWeights(v.To, weights); err != nil {
		log.Printf("%s: write data error: %v\n", name, err)
		return nil
	}

	log.Printf("%s: OK\n", name)
	return nil
}

type syncConfig struct {
	From any               `yaml:"from"`
	To   string            `yaml:"to"`
	Expr map[string]string `yaml:"expr"`
}

func countSyncs(data []byte) (int, error) {
	var syncs map[string]syncConfig
	if err := yaml.Unmarshal(data, &syncs); err != nil {
		return 0, err
	}

	total := 0
	for _, v := range syncs {
		if v.From == "" || v.To == "" {
			continue
		}
		total++
	}
	return total, nil
}

type processRequest struct {
	Data   []byte
	Name   string
	Config syncConfig
}

func setupHAMQTTResets(data []byte, requests chan<- processRequest) ([]func(), error) {
	if data == nil {
		return nil, nil
	}

	var syncs map[string]syncConfig
	if err := yaml.Unmarshal(data, &syncs); err != nil {
		return nil, err
	}

	var closers []func()
	for name, v := range syncs {
		if v.From == "" || v.To == "" || !internal.IsHAMQTT(v.To) {
			continue
		}

		syncID, err := getSyncID(name, v)
		if err != nil {
			return closers, fmt.Errorf("%s: %w", name, err)
		}

		name := name
		config := v
		close, err := internal.SubscribeHAMQTTReset(v.To, name, syncID, func() {
			select {
			case requests <- processRequest{Name: name, Config: config}:
				log.Printf("%s: queued by mqtt reset\n", name)
			default:
				log.Printf("%s: mqtt reset skipped because queue is full\n", name)
			}
		})
		if err != nil {
			return closers, fmt.Errorf("%s: %w", name, err)
		}

		closers = append(closers, close)
		log.Printf("%s: ha_mqtt reset subscribed\n", name)
	}

	return closers, nil
}

func normalizeDBPath(path string) (string, error) {
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = wd
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "scaleconnect.db")
	} else if filepath.Ext(path) == "" {
		path = filepath.Join(path, "scaleconnect.db")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func getSyncID(name string, config any) (string, error) {
	data, err := json.Marshal(struct {
		Name   string `json:"name"`
		Config any    `json:"config"`
	}{
		Name:   name,
		Config: config,
	})
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:16]), nil
}

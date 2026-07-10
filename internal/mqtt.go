package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

const TargetHAMQTT = "ha_mqtt"

type HAMQTTConfig struct {
	Broker      string
	Username    string
	Password    string
	ConfigTopic string
	ResetTopic  string
}

var mqttDiscoveryIDRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
var mqttDiscoveryPublished sync.Map

func IsHAMQTT(config string) bool {
	fields := strings.Fields(config)
	return len(fields) > 0 && fields[0] == TargetHAMQTT
}

func ParseHAMQTT(config string) (HAMQTTConfig, error) {
	fields := strings.Fields(config)
	if len(fields) < 6 || fields[0] != TargetHAMQTT {
		return HAMQTTConfig{}, errors.New("ha_mqtt format: ha_mqtt {ha_addr} {username} {password} {chconfig} {chreset}")
	}

	broker, err := normalizeMQTTBroker(fields[1])
	if err != nil {
		return HAMQTTConfig{}, err
	}

	return HAMQTTConfig{
		Broker:      broker,
		Username:    fields[2],
		Password:    fields[3],
		ConfigTopic: fields[4],
		ResetTopic:  fields[5],
	}, nil
}

func normalizeMQTTBroker(addr string) (string, error) {
	if addr == "" {
		return "", errors.New("empty mqtt broker address")
	}
	if strings.HasPrefix(addr, "mqtt://") {
		addr = "tcp://" + strings.TrimPrefix(addr, "mqtt://")
	}
	if !strings.Contains(addr, "://") {
		addr = "tcp://" + addr
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid mqtt broker address: %s", addr)
	}
	if u.Port() == "" {
		u.Host = net.JoinHostPort(u.Hostname(), "1883")
	}

	return u.String(), nil
}

func PublishHAMQTT(config, name, syncID string, weights []*core.Weight) error {
	cfg, err := ParseHAMQTT(config)
	if err != nil {
		return err
	}

	client := mqtt.NewClient(mqttOptions(cfg, "pub"))
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer client.Disconnect(250)

	if err = publishHAMQTTDiscovery(client, cfg, name, syncID); err != nil {
		return err
	}

	dst := append([]*core.Weight(nil), weights...)
	slices.SortFunc(dst, func(a, b *core.Weight) int {
		return a.Date.Compare(b.Date)
	})

	for _, weight := range dst {
		data, err := json.Marshal(weight)
		if err != nil {
			return err
		}

		stateTopic := cfg.StateTopic(syncID)
		token := client.Publish(stateTopic, 0, true, data)
		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
		log.Printf("ha_mqtt state published topic=%s retain=true date=%s\n", stateTopic, weight.Date.Format(time.RFC3339))
	}

	return nil
}

func SubscribeHAMQTTReset(config, name, syncID string, onReset func()) (func(), error) {
	cfg, err := ParseHAMQTT(config)
	if err != nil {
		return nil, err
	}
	if cfg.ResetTopic == "" {
		return func() {}, nil
	}

	client := mqtt.NewClient(mqttOptions(cfg, "sub"))
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	if err = publishHAMQTTDiscovery(client, cfg, name, syncID); err != nil {
		client.Disconnect(250)
		return nil, err
	}

	token := client.Subscribe(cfg.ResetTopic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		if len(msg.Payload()) == 0 {
			return
		}
		log.Printf("ha_mqtt reset topic=%s payload=%s\n", msg.Topic(), string(msg.Payload()))
		onReset()
	})
	if token.Wait() && token.Error() != nil {
		client.Disconnect(250)
		return nil, token.Error()
	}

	return func() {
		client.Disconnect(250)
	}, nil
}

func mqttOptions(cfg HAMQTTConfig, suffix string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(fmt.Sprintf("scaleconnect-%s-%d", suffix, time.Now().UnixNano()))
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetAutoReconnect(true)
	return opts
}

func publishHAMQTTDiscovery(client mqtt.Client, cfg HAMQTTConfig, name, syncID string) error {
	cacheKey := cfg.ConfigTopic + ":" + syncID
	if _, loaded := mqttDiscoveryPublished.LoadOrStore(cacheKey, struct{}{}); loaded {
		return nil
	}

	objectID := "scaleconnect_" + mqttDiscoveryID(syncID)
	shortID := mqttShortID(syncID)
	stateTopic := cfg.StateTopic(syncID)
	payload := map[string]any{
		"name":                  name + " " + shortID,
		"unique_id":             objectID + "_weight",
		"state_topic":           stateTopic,
		"value_template":        "{{ value_json.Weight }}",
		"json_attributes_topic": stateTopic,
		"device_class":          "weight",
		"state_class":           "measurement",
		"unit_of_measurement":   "kg",
		"force_update":          true,
		"device": map[string]any{
			"identifiers":  []string{objectID},
			"name":         "ScaleConnect " + name + " " + shortID,
			"manufacturer": "SmartScaleConnect",
			"model":        "Smart Scale Sync",
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := client.Publish(cfg.ConfigTopic, 0, true, data)
	if token.Wait() && token.Error() != nil {
		mqttDiscoveryPublished.Delete(cacheKey)
		return token.Error()
	}
	log.Printf("ha_mqtt config published topic=%s retain=true state_topic=%s\n", cfg.ConfigTopic, stateTopic)

	return nil
}

func (cfg HAMQTTConfig) StateTopic(syncID string) string {
	prefix := strings.TrimSuffix(cfg.ConfigTopic, "/config")
	if prefix == cfg.ConfigTopic {
		return strings.TrimSuffix(cfg.ConfigTopic, "/") + "/state/" + mqttDiscoveryID(syncID)
	}
	return prefix + "/state/" + mqttDiscoveryID(syncID)
}

func mqttDiscoveryID(value string) string {
	value = mqttDiscoveryIDRe.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "default"
	}
	return strings.ToLower(value)
}

func mqttShortID(value string) string {
	value = mqttDiscoveryID(value)
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

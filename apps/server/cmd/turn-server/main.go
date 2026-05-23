package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pion/turn/v4"
)

type turnServerConfig struct {
	ListenIP     string
	PublicIP     string
	Port         int
	Realm        string
	Username     string
	Password     string
	RelayMinPort int
	RelayMaxPort int
}

func main() {
	cfg := parseConfig()
	if err := validateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	resolvedPublicIP, err := resolvePublicIP(cfg.PublicIP)
	if err != nil {
		log.Fatalf("resolve public IP failed: %v", err)
	}

	udpListener, err := net.ListenPacket("udp4", net.JoinHostPort(cfg.ListenIP, strconv.Itoa(cfg.Port)))
	if err != nil {
		log.Fatalf("TURN UDP listen failed: %v", err)
	}

	tcpListener, err := net.Listen("tcp4", net.JoinHostPort(cfg.ListenIP, strconv.Itoa(cfg.Port)))
	if err != nil {
		_ = udpListener.Close()
		log.Fatalf("TURN TCP listen failed: %v", err)
	}

	relayAddressGenerator := &turn.RelayAddressGeneratorPortRange{
		RelayAddress: resolvedPublicIP,
		Address:      cfg.ListenIP,
		MinPort:      uint16(cfg.RelayMinPort),
		MaxPort:      uint16(cfg.RelayMaxPort),
		MaxRetries:   50,
	}

	userKey := turn.GenerateAuthKey(cfg.Username, cfg.Realm, cfg.Password)
	server, err := turn.NewServer(turn.ServerConfig{
		Realm: cfg.Realm,
		AuthHandler: func(username string, realm string, sourceAddress net.Addr) ([]byte, bool) {
			if username != cfg.Username || realm != cfg.Realm {
				log.Printf("TURN auth rejected username=%s realm=%s source=%s", username, realm, sourceAddress.String())
				return nil, false
			}
			return userKey, true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn:            udpListener,
				RelayAddressGenerator: relayAddressGenerator,
			},
		},
		ListenerConfigs: []turn.ListenerConfig{
			{
				Listener:              tcpListener,
				RelayAddressGenerator: relayAddressGenerator,
			},
		},
	})
	if err != nil {
		_ = udpListener.Close()
		_ = tcpListener.Close()
		log.Fatalf("TURN server init failed: %v", err)
	}

	log.Printf(
		"TURN server listening on %s, advertised as %s:%d, relay ports %d-%d, realm=%s, username=%s",
		net.JoinHostPort(cfg.ListenIP, strconv.Itoa(cfg.Port)),
		resolvedPublicIP.String(),
		cfg.Port,
		cfg.RelayMinPort,
		cfg.RelayMaxPort,
		cfg.Realm,
		cfg.Username,
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := server.Close(); err != nil {
		log.Printf("TURN server close failed: %v", err)
	}
}

func parseConfig() turnServerConfig {
	var cfg turnServerConfig
	flag.StringVar(&cfg.ListenIP, "listen-ip", getEnv("TURN_LISTEN_IP", "0.0.0.0"), "IP address to bind TURN listeners")
	flag.StringVar(&cfg.PublicIP, "public-ip", getEnv("TURN_PUBLIC_IP", "auto"), "IP address advertised in relay candidates, or auto")
	flag.IntVar(&cfg.Port, "port", getIntEnv("TURN_HOST_PORT", 3478), "TURN listener port")
	flag.StringVar(&cfg.Realm, "realm", getEnv("TURN_REALM", "robot-center.local"), "TURN realm")
	flag.StringVar(&cfg.Username, "username", getEnv("TURN_USERNAME", "robot"), "TURN username")
	flag.StringVar(&cfg.Password, "password", getEnv("TURN_PASSWORD", "robot-pass"), "TURN password")
	flag.IntVar(&cfg.RelayMinPort, "relay-min-port", getIntEnv("TURN_RELAY_MIN_PORT", 49160), "minimum relay port")
	flag.IntVar(&cfg.RelayMaxPort, "relay-max-port", getIntEnv("TURN_RELAY_MAX_PORT", 49180), "maximum relay port")
	flag.Parse()
	return cfg
}

func validateConfig(cfg turnServerConfig) error {
	if cfg.ListenIP == "" {
		return errors.New("listen IP is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid TURN port: %d", cfg.Port)
	}
	if cfg.Realm == "" {
		return errors.New("realm is required")
	}
	if cfg.Username == "" {
		return errors.New("username is required")
	}
	if cfg.Password == "" {
		return errors.New("password is required")
	}
	if cfg.RelayMinPort < 1 || cfg.RelayMinPort > 65535 {
		return fmt.Errorf("invalid relay min port: %d", cfg.RelayMinPort)
	}
	if cfg.RelayMaxPort < 1 || cfg.RelayMaxPort > 65535 {
		return fmt.Errorf("invalid relay max port: %d", cfg.RelayMaxPort)
	}
	if cfg.RelayMinPort > cfg.RelayMaxPort {
		return errors.New("relay min port must be lower than or equal to relay max port")
	}
	return nil
}

func resolvePublicIP(value string) (net.IP, error) {
	if value != "" && value != "auto" {
		parsedIP := net.ParseIP(value)
		if parsedIP == nil {
			return nil, fmt.Errorf("invalid public IP: %s", value)
		}
		return parsedIP, nil
	}

	if ip := discoverOutboundIP(); ip != nil {
		return ip, nil
	}
	if ip := findFirstInterfaceIP(); ip != nil {
		return ip, nil
	}
	return nil, errors.New("public IP auto detection failed; set TURN_PUBLIC_IP explicitly")
}

func discoverOutboundIP() net.IP {
	connection, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer connection.Close()

	localAddress, ok := connection.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil
	}
	return localAddress.IP.To4()
}

func findFirstInterfaceIP() net.IP {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ipNet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}
			return ip
		}
	}
	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsedValue
}

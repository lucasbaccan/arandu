package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"devopsconecta/backend/internal/config"
)

// resolveConfig decide a configuração na ordem: flags > -defaults > variáveis de ambiente > menu interativo > defaults.
func resolveConfig() config.Config {
	cfg := config.Load()

	host := flag.String("host", "", "host do servidor (ex: 0.0.0.0)")
	port := flag.String("port", "", "porta do servidor (ex: 8080)")
	db := flag.String("db", "", "caminho do banco SQLite (ex: ./data/app.db)")
	useDefaults := flag.Bool("yes", false, "assume todos os valores padrão, sem menu interativo")
	flag.BoolVar(useDefaults, "defaults", false, "mesma coisa que --yes")
	flag.Parse()

	hasFlag := flag.NFlag() > 0
	envSet := func(k string) bool { return os.Getenv(k) != "" }

	if *useDefaults {
		cfg.Host = config.DefaultHost
		cfg.Port = config.DefaultPort
		cfg.DatabasePath = config.DefaultDB
	} else if !hasFlag && !envSet("HOST") && !envSet("PORT") && !envSet("DATABASE_PATH") {
		cfg = promptSetup(cfg)
	}

	if *host != "" {
		cfg.Host = *host
	}
	if *port != "" {
		cfg.Port = *port
	}
	if *db != "" {
		cfg.DatabasePath = *db
	}
	return cfg
}

// promptSetup pergunta host, porta e banco quando nada foi configurado.
// Em ambientes não interativos (stdin sem terminal), usa os defaults sem perguntar.
func promptSetup(cfg config.Config) config.Config {
	if !isTerminal(os.Stdin) {
		fmt.Println("Nenhuma variável de ambiente configurada; usando valores padrão.")
		fmt.Printf("Host: %s | Porta: %s | Banco: %s\n", cfg.Host, cfg.Port, cfg.DatabasePath)
		fmt.Println("Dica: configure HOST, PORT e DATABASE_PATH para silenciar esta mensagem.")
		return cfg
	}

	fmt.Println("===== DevOps Conecta — configuração inicial =====")
	fmt.Println("Pressione Enter para aceitar o valor padrão.")
	cfg.Host = ask("Host", cfg.Host)
	cfg.Port = ask("Porta", cfg.Port)
	cfg.DatabasePath = ask("Banco de dados (SQLite)", cfg.DatabasePath)
	fmt.Printf("===== Iniciando em http://%s:%s =====\n", cfg.Host, cfg.Port)
	return cfg
}

func ask(label, def string) string {
	fmt.Printf("%s [%s]: ", label, def)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

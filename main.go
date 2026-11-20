package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	downloadURL = "https://raw.githubusercontent.com/sebaxakerhtc/rdpwrap.ini/master/rdpwrap.ini"
	installDir  = `C:\Program Files\RDP Wrapper`
	fileName    = "rdpwrap.ini"
)

func main() {
	fmt.Println("===================================")
	fmt.Println("     Atualizador RDP Wrapper v1.0 by Faezito")
	fmt.Println("===================================")
	fmt.Println()

	if !isAdmin() {
		fmt.Println("ERRO: Execute este programa como Administrador.")
		pause()
		return
	}

	target := filepath.Join(installDir, fileName)

	fmt.Println("[1/6] Parando Serviços de Área de Trabalho Remota")

	_ = runCommand("net", "stop", "TermService")

	fmt.Println("[2/6] Baixando arquivo atualizado...")

	data, err := download()
	if err != nil {
		fail(err)
		return
	}

	fmt.Printf("      Download concluido: %d bytes\n\n", len(data))

	fmt.Println("[3/6] Validando arquivo...")

	if len(data) == 0 ||
		!strings.Contains(string(data), "[PatchCodes]") {
		fail(fmt.Errorf("arquivo invalido ou incompleto"))
		return
	}

	fmt.Println("      Arquivo validado.\n")

	fmt.Println("[4/6] Criando backup...")

	backup := target + "." +
		time.Now().Format("20060102_150405") + ".bak"

	if _, err := os.Stat(target); err != nil {
		fail(fmt.Errorf("arquivo original nao encontrado: %w", err))
		return
	}

	if err := os.Rename(target, backup); err != nil {
		fail(fmt.Errorf("nao foi possivel criar backup: %w", err))
		return
	}

	fmt.Println("      Backup:", backup)

	fmt.Println("\n[5/6] Instalando novo arquivo...")

	tmp, err := os.CreateTemp(installDir, "rdpwrap-*.tmp")
	if err != nil {
		_ = os.Rename(backup, target)
		fail(err)
		return
	}

	tmpName := tmp.Name()

	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()

	if writeErr != nil || closeErr != nil {
		_ = os.Remove(tmpName)
		_ = os.Rename(backup, target)

		if writeErr != nil {
			fail(writeErr)
		} else {
			fail(closeErr)
		}
		return
	}

	if err := os.Rename(tmpName, target); err != nil {
		_ = os.Remove(tmpName)
		_ = os.Rename(backup, target)
		fail(err)
		return
	}

	fmt.Println("      Arquivo substituido com sucesso.")

	fmt.Println("\n[6/6] Reiniciando servico TermService...")

	stopErr := runCommand("net", "stop", "TermService")
	if stopErr != nil {
		fmt.Println("      Aviso ao parar servico:", stopErr)
	}

	startErr := runCommand("net", "start", "TermService")
	if startErr != nil {
		fail(fmt.Errorf("arquivo atualizado, mas nao foi possivel iniciar o servico: %w", startErr))
		return
	}

	fmt.Println("\n===================================")
	fmt.Println("       ATUALIZACAO CONCLUIDA")
	fmt.Println("===================================")
	fmt.Println("Arquivo:", target)
	fmt.Println("Backup: ", backup)

	pause()
}

func download() ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro HTTP: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isAdmin() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("IsUserAnAdmin")

	result, _, _ := proc.Call()
	return result != 0
}

func fail(err error) {
	fmt.Println("\nERRO:", err)
	_ = runCommand("net", "start", "TermService")
	pause()
}

func pause() {
	fmt.Println("\nPressione ENTER para sair...")
	fmt.Scanln()
}

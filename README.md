# Atualizador RDP Wrapper

Utilitário de linha de comando (CLI) em **Go** para automatizar a atualização do arquivo `rdpwrap.ini` no Windows, realizando validação básica do conteúdo, backup automático com timestamp, escrita segura via arquivo temporário e reinicialização do serviço de Área de Trabalho Remota (`TermService`).

---

## Como funciona

O programa executa um fluxo em **6 etapas** para reduzir o risco de deixar o `rdpwrap.ini` corrompido ou o serviço parado:

1. **Verificação de privilégios:** Confere se o processo está rodando como **Administrador** (via `IsUserAnAdmin` na `shell32.dll`).
2. **Parada preventiva do serviço:** Tenta interromper o `TermService` (`net stop TermService`) antes de manipular o arquivo, a fim de evitar erros de escrita causado pelo arquivo em execução.
3. **Download do `rdpwrap.ini`:** Baixa a versão mais recente a partir do repositório da comunidade (`sebaxakerhtc/rdpwrap.ini`, branch `master`), com timeout de 30 segundos e verificação de status `HTTP 200 OK`.
4. **Validação do conteúdo:** Garante que o arquivo baixado não está vazio e contém a seção `[PatchCodes]`.
5. **Backup com timestamp:** Renomeia o `rdpwrap.ini` atual em `C:\Program Files\RDP Wrapper` para `rdpwrap.ini.AAAAMMDD_HHMMSS.bak` (ex.: `rdpwrap.ini.20261120_163000.bak`).
6. **Instalação atômica e reinício do serviço:**
   - Grava o novo conteúdo em um arquivo temporário (`rdpwrap-*.tmp`) dentro da mesma pasta de destino.
   - Em caso de falha na gravação, fechamento ou substituição, remove o temporário e restaura automaticamente o `.bak` para `rdpwrap.ini`.
   - Reinicia o `TermService` (`net stop TermService` seguido de `net start TermService`).

> **Recuperação em caso de erro:** Se ocorrer alguma falha durante o processo, a função de erro tenta subir novamente o `TermService` (`net start TermService`) antes de aguardar `ENTER` para encerrar o console.

---

## Pré-requisitos e caminhos

- **Sistema operacional:** Windows.
- **Permissões:** Execução obrigatória **Como Administrador**.
- **Diretório de instalação esperado:** `C:\Program Files\RDP Wrapper\rdpwrap.ini` (o `rdpwrap.ini` original já precisa existir nesse caminho no momento do backup).
- **Conexão com a internet:** Necessária para baixar o `rdpwrap.ini` do GitHub.

---

## Como usar (via Release `.exe`)

1. Acesse a aba **Releases** deste repositório e baixe o arquivo `.exe` mais recente.
2. Clique com o botão direito sobre o `.exe` e selecione **Executar como administrador**.
3. Acompanhe o progresso das etapas `[1/6]` a `[6/6]` no terminal.
4. Ao final, pressione `ENTER` para fechar a janela.

> **Aviso do Windows / SmartScreen:** Como binários independentes sem assinatura digital podem acionar o Windows SmartScreen ou antivírus, caso apareça um aviso, clique em **Mais informações** > **Executar assim mesmo** (ou valide o hash do arquivo publicado na release).

---

## Compilando a partir do código-fonte

Caso queira compilar o binário manualmente com o Go instalado:

```bash
go build -ldflags="-s -w" -o RDPWrapperUpdater.exe .
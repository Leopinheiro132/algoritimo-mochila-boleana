package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed pecas.csv
var pecasCSVPadrao []byte

type Peca struct {
	Nome          string
	Preco         int
	kmUltimatroca int
	vidaUtil      int
	criticidade   int
}

func (p Peca) Desgaste(kmAtual int) float64 {
	d := float64(kmAtual-p.kmUltimatroca) / float64(p.vidaUtil)
	if d > 1.0 {
		d = 1.0
	}
	return d
}

func (p Peca) Valor(kmAtual int) int {
	d := p.Desgaste(kmAtual)
	v := int(d*100) * p.criticidade
	if d >= 1.0 {
		v += 100 * p.criticidade
	}
	return v
}

func resolverArquivo(caminho string) (io.ReadCloser, error) {
	// 1. Tenta abrir no caminho especificado diretamente
	if f, err := os.Open(caminho); err == nil {
		return f, nil
	}

	// 2. Se executado a partir da raiz do projeto, tenta dentro de algoritimo/
	caminhoNaPasta := filepath.Join("algoritimo", caminho)
	if f, err := os.Open(caminhoNaPasta); err == nil {
		return f, nil
	}

	// 3. Fallback: se for o arquivo padrão 'pecas.csv', usa os dados embutidos (embed)
	if (caminho == "pecas.csv" || filepath.Base(caminho) == "pecas.csv") && len(pecasCSVPadrao) > 0 {
		return io.NopCloser(bytes.NewReader(pecasCSVPadrao)), nil
	}

	return nil, fmt.Errorf("arquivo %q não encontrado", caminho)
}

func LerPecas(caminho string) ([]Peca, error) {
	leitor, err := resolverArquivo(caminho)
	if err != nil {
		return nil, err
	}
	defer leitor.Close()

	br := bufio.NewReader(leitor)

	// Remove BOM (caso o arquivo tenha sido salvo no Windows/Excel com UTF-8 BOM)
	if b, _ := br.Peek(3); len(b) == 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		br.Discard(3)
	}

	r := csv.NewReader(br)
	r.TrimLeadingSpace = true

	// Lê todas as linhas do CSV
	linhas, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("não consegui ler o CSV: %w", err)
	}

	if len(linhas) <= 1 {
		return nil, errors.New("o arquivo CSV está vazio ou contém apenas o cabeçalho")
	}

	// Descobre o índice de cada coluna pelo cabeçalho (linha 0)
	colNome, colPreco, colKm, colVida, colCrit := -1, -1, -1, -1, -1
	for i, col := range linhas[0] {
		c := strings.ToLower(strings.TrimSpace(col))
		switch c {
		case "nome", "peca", "item":
			colNome = i
		case "preco", "valor", "custo":
			colPreco = i
		case "km_usada", "km_ultima_troca", "km_troca", "km_ultima":
			colKm = i
		case "vida_util_km", "vida_util", "intervalo":
			colVida = i
		case "importancia", "criticidade", "prioridade":
			colCrit = i
		}
	}

	if colNome == -1 || colPreco == -1 || colKm == -1 || colVida == -1 || colCrit == -1 {
		return nil, fmt.Errorf("colunas necessárias não encontradas no cabeçalho: %v", linhas[0])
	}

	var pecas []Peca

	// Itera a partir da linha 1 (pulando o cabeçalho)
	for i, reg := range linhas[1:] {
		linhaNum := i + 2

		// Pula linha em branco
		if len(reg) == 0 || (len(reg) == 1 && strings.TrimSpace(reg[0]) == "") {
			continue
		}

		nome := strings.TrimSpace(reg[colNome])
		if nome == "" {
			return nil, fmt.Errorf("linha %d: nome da peça está vazio", linhaNum)
		}

		preco, err := strconv.Atoi(strings.TrimSpace(reg[colPreco]))
		if err != nil {
			return nil, fmt.Errorf("linha %d: preço inválido (%q)", linhaNum, reg[colPreco])
		}

		kmTroca, err := strconv.Atoi(strings.TrimSpace(reg[colKm]))
		if err != nil {
			return nil, fmt.Errorf("linha %d: km inválido (%q)", linhaNum, reg[colKm])
		}

		vidaUtil, err := strconv.Atoi(strings.TrimSpace(reg[colVida]))
		if err != nil {
			return nil, fmt.Errorf("linha %d: vida útil inválida (%q)", linhaNum, reg[colVida])
		}

		crit, err := strconv.Atoi(strings.TrimSpace(reg[colCrit]))
		if err != nil {
			return nil, fmt.Errorf("linha %d: importância/criticidade inválida (%q)", linhaNum, reg[colCrit])
		}

		pecas = append(pecas, Peca{
			Nome:          nome,
			Preco:         preco,
			kmUltimatroca: kmTroca,
			vidaUtil:      vidaUtil,
			criticidade:   crit,
		})
	}

	return pecas, nil
}

func Mochila(pecas []Peca, orcamento int, kmAtual int) ([]Peca, int) {
	n := len(pecas)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, orcamento+1)
	}

	for i := 1; i <= n; i++ {
		p := pecas[i-1]
		v := p.Valor(kmAtual)
		for w := 0; w <= orcamento; w++ {
			dp[i][w] = dp[i-1][w] // não inclui a peça
			if p.Preco <= w {
				if cand := dp[i-1][w-p.Preco] + v; cand > dp[i][w] {
					dp[i][w] = cand // pega a peça
				}
			}
		}
	}

	var escolhidas []Peca
	w := orcamento
	for i := n; i > 0 && w > 0; i-- {
		if dp[i][w] != dp[i-1][w] {
			escolhidas = append(escolhidas, pecas[i-1])
			w -= pecas[i-1].Preco
		}
	}

	sort.Slice(escolhidas, func(a, b int) bool {
		return escolhidas[a].Valor(kmAtual) > escolhidas[b].Valor(kmAtual)
	})
	return escolhidas, dp[n][orcamento]
}

// validação em força bruta
func ForcaBruta(pecas []Peca, orcamento int, kmAtual int) int {
	n := len(pecas)
	valores := make([]int, n)
	for i, p := range pecas {
		valores[i] = p.Valor(kmAtual)
	}
	melhor := 0
	for mascara := uint(0); mascara < (1 << uint(n)); mascara++ {
		custo, valor := 0, 0
		for i := 0; i < n; i++ {
			if mascara&(1<<uint(i)) != 0 {
				custo += pecas[i].Preco
				valor += valores[i]
			}
		}
		if custo <= orcamento && valor > melhor {
			melhor = valor
		}
	}
	return melhor
}

func gerarCenario(rng *rand.Rand, n int) ([]Peca, int, int) {
	kmAtual := 10000 + rng.Intn(40000)
	pecas := make([]Peca, n)
	soma := 0
	for i := range pecas {
		pecas[i] = Peca{
			Nome:          string(rune('A' + i)),
			Preco:         10 + rng.Intn(400),
			kmUltimatroca: kmAtual - (rng.Intn(5) * 20000),
			vidaUtil:      5000 + rng.Intn(20000),
			criticidade:   1 + rng.Intn(10),
		}
		soma += pecas[i].Preco
	}
	return pecas, 1 + rng.Intn(soma), kmAtual
}

func Validar(casos int) bool {
	rng := rand.New(rand.NewSource(42))
	for c := 0; c < casos; c++ {
		n := 1 + rng.Intn(14)
		pecas, orcamento, km := gerarCenario(rng, n)

		escolhidas, valorDP := Mochila(pecas, orcamento, km)
		valorBF := ForcaBruta(pecas, orcamento, km)

		if valorDP != valorBF {
			fmt.Printf("FALHA caso %d: DP=%d BF=%d\n", c, valorDP, valorBF)
			return false
		}

		custo, soma := 0, 0
		vistos := make(map[string]bool)
		for _, p := range escolhidas {
			if vistos[p.Nome] {
				fmt.Printf("FALHA caso %d: peça %s repetida\n", c, p.Nome)
				return false
			}
			vistos[p.Nome] = true
			custo += p.Preco
			soma += p.Valor(km)
		}
		if custo > orcamento || soma != valorDP {
			fmt.Printf("FALHA caso %d: custo=%d orçamento=%d soma=%d dp=%d\n",
				c, custo, orcamento, soma, valorDP)
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------

func main() {
	arquivo := "pecas.csv"
	km := 32000
	orcamento := 600

	pecas, err := LerPecas(arquivo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler %s: %v\n", arquivo, err)
		os.Exit(1)
	}

	escolhidas, total := Mochila(pecas, orcamento, km)

	fmt.Printf("Arquivo: %s (%d peças) | Km atual: %d | Orçamento: R$ %d\n\n",
		arquivo, len(pecas), km, orcamento)

	fmt.Println("Comprar/trocar agora (ordem de prioridade):")
	gasto := 0
	selecionada := make(map[string]bool, len(escolhidas))
	for _, p := range escolhidas {
		fmt.Printf("- %-38s R$ %5d | desgaste %3.0f%% | urgência %d\n",
			p.Nome, p.Preco, p.Desgaste(km)*100, p.Valor(km))
		gasto += p.Preco
		selecionada[p.Nome] = true
	}
	fmt.Printf("\nTotal gasto: R$ %d | Sobra: R$ %d | Urgência total: %d\n",
		gasto, orcamento-gasto, total)

	// Peças que não couberam, da mais urgente para a menos urgente
	var fora []Peca
	for _, p := range pecas {
		if !selecionada[p.Nome] {
			fora = append(fora, p)
		}
	}
	sort.Slice(fora, func(a, b int) bool { return fora[a].Valor(km) > fora[b].Valor(km) })
	if len(fora) > 0 {
		fmt.Println("\nFicaram de fora (próximo ciclo):")
		for _, p := range fora {
			fmt.Printf("- %-38s R$ %5d | desgaste %3.0f%% | urgência %d\n",
				p.Nome, p.Preco, p.Desgaste(km)*100, p.Valor(km))
		}
	}

	fmt.Println("\nValidando DP contra força bruta (500 cenários)...")
	if Validar(500) {
		fmt.Println("OK: DP igual à força bruta em todos os cenários.")
	}
}

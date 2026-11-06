# Guia Completo: Otimização de Troca de Peças da Moto

Este documento foi feito para quem está **aprendendo do zero** como funcionam os algoritmos deste projeto. Vamos explicar passo a passo o problema da moto, o funcionamento da **Mochila (Programação Dinâmica)** e o sistema de **Validação contra Força Bruta**.

---

## 1. O Problema da Vida Real

Imagine o seguinte cenário:
Você tem uma moto com **32.000 km**. Você vai à oficina e o mecânico lista várias peças desgastadas. 
* Se você fosse trocar **tudo**, gastaria mais de R$ 3.000.
* Porém, você só tem **R$ 600 no bolso**.

**A pergunta de ouro:** *Quais peças você deve escolher trocar HOJE para ter a maior segurança e evitar que a moto quebre, sem passar dos seus R$ 600?*

Na ciência da computação, esse dilema é idêntico ao famoso **Problema da Mochila 0/1 (*0/1 Knapsack Problem*)**:
* A **mochila** é o seu bolso (capacidade = orçamento em R$).
* Os **itens** são as peças da moto.
* O **peso** do item é o seu preço em R$.
* O **benefício** do item é a urgência/risco da peça quebrar.
* **0/1** significa: você não pode trocar "meia pastilha"; ou você troca a peça inteira (**1**), ou não troca nada (**0**).

---

## 2. Como medimos a "Urgência" de uma Peça?

Antes de escolher o que comprar, precisamos dar uma nota matemática para o perigo de cada peça. No arquivo [`main.go`](main.go), isso é feito em duas etapas:

### Passo A: Desgaste (`Desgaste`)
```go
func (p Peca) Desgaste(kmAtual int) float64 {
	d := float64(kmAtual - p.kmUltimatroca) / float64(p.vidaUtil)
	if d > 1.0 {
		d = 1.0 // no máximo 100% de desgaste
	}
	return d
}
```
* **O que faz:** Calcula quantos porcento da vida útil já foram consumidos.
* **Exemplo:** Se uma pastilha dura `15.000 km` e você rodou `15.000 km` com ela, $15.000 / 15.000 = 1.0$ ($100\%$ gasta).

### Passo B: Pontuação de Valor (`Valor`)
```go
func (p Peca) Valor(kmAtual int) int {
	d := p.Desgaste(kmAtual)
	v := int(d*100) * p.criticidade
	if d >= 1.0 {
		v += 100 * p.criticidade // bônus de perigo se já passou da hora de trocar!
	}
	return v
}
```
* Multiplicamos o desgaste pela **criticidade** (importância de 1 a 10).
* Se a peça atingiu $100\%$ do desgaste, dobramos os pontos de perigo. Uma pastilha de freio careca é prioridade máxima em relação a um espelho retrovisor trincado!

---

## 3. A Mochila em Si (`Mochila` - Programação Dinâmica)

A função `Mochila` resolve esse problema de forma inteligente sem precisar testar combinações às cegas.

### A Tabela no Caderno (`dp[i][w]`)
Imagine uma tabela desenhada num caderno:
* **Linhas (`i`)**: Representam as peças que vamos tirando da caixa, uma por uma (de 1 até $n$).
* **Colunas (`w`)**: Representam orçamentos hipotéticos menores, de R$ 0 até o seu limite (ex: R$ 0, R$ 1, R$ 2 ... R$ 600).
* **Célula `dp[i][w]`**: Guarda a **maior pontuação de urgência possível** considerando as primeiras `i` peças, com um teto de gastos de `w` reais.

```
       w = R$ 0   R$ 1   ...   R$ 80   ...   R$ 600 (Orçamento total)
i = 0 [   0        0     ...     0     ...     0   ] (Nenhuma peça avaliada)
i = 1 [   0        0     ...   2000    ...   2000  ] (Apenas Pastilha de Freio)
i = 2 [   ...                                      ]
```

### O Dilema em Cada Casa: Levar ou Não Levar?

Para cada peça e para cada valor de orçamento `w`, o computador faz uma pergunta muito simples:

```go
for i := 1; i <= n; i++ {
    p := pecas[i-1]
    v := p.Valor(kmAtual)
    for w := 0; w <= orcamento; w++ {
        
        // 1ª Opção: NÃO LEVAR essa peça
        dp[i][w] = dp[i-1][w]

        // 2ª Opção: Tentar LEVAR essa peça
        if p.Preco <= w {
            // Se eu pagar o preço dela, quanto dinheiro me sobra? (w - p.Preco)
            // Qual era a melhor pontuação que eu tinha com esse dinheiro que sobrou?
            candidato := dp[i-1][w - p.Preco] + v

            // Compensa levar?
            if candidato > dp[i][w] {
                dp[i][w] = candidato
            }
        }
    }
}
```

* **Não Levar:** Você simplesmente copia o melhor valor que já tinha conseguido sem essa peça na linha de cima: `dp[i-1][w]`.
* **Levar:** Se a peça couber no bolso atual (`p.Preco <= w`):
  * Você ganha o valor `v` dela.
  * Mas seu dinheiro cai para `w - p.Preco`.
  * Você consulta a linha anterior para ver o que dava para fazer com o dinheiro que sobrou: `dp[i-1][w - p.Preco]`.
  * Se a soma for maior do que não levar, você escolhe levar!

### Reconstrução: Quais peças foram escolhidas?
Ao final dos loops, a célula `dp[n][orcamento]` sabe a **pontuação máxima**, mas não sabe os nomes das peças.
Para descobrir os nomes, o código faz o caminho inverso ("volta no tempo"):

```go
var escolhidas []Peca
w := orcamento
for i := n; i > 0 && w > 0; i-- {
    // Se o valor mudou em relação à linha anterior, significa que essa peça foi comprada!
    if dp[i][w] != dp[i-1][w] {
        escolhidas = append(escolhidas, pecas[i-1])
        w -= pecas[i-1].Preco // desconta o preço da peça do dinheiro restante
    }
}
```

---

## 4. O Sistema de Validação: Como ter certeza de que o código não errou?

Em sistemas críticos ou algoritmos complexos, **como saber se a Programação Dinâmica realmente encontrou a resposta perfeita?**

Para isso, usamos uma técnica chamada **Validação Cruzada com Oráculo**:
Criamos um segundo algoritmo muito mais simples e "burro", que testa **todas as combinações do universo**. Se a Programação Dinâmica der o mesmo resultado que o teste exaustivo em 500 cenários aleatórios diferentes, temos a certeza absoluta de que o código está certo.

```
       [ Gerador de Cenários Aleatórios ]
                       │
       ┌───────────────┴───────────────┐
       ▼                               ▼
[ Algoritmo Mochila (DP) ]    [ Força Bruta (Bitmask) ]
       │                               │
       └───────────────┬───────────────┘
                       ▼
       ¿ Os dois resultados são iguais ?
            Sim ──► APROVADO (OK)
            Não ──► FALHA NO ALGORITMO
```

---

### A Força Bruta com Máscara de Bits (`ForcaBruta`)

A Força Bruta não tenta ser rápida; ela tenta ser **infalível**. Ela testa todas as $2^n$ combinações possíveis.

#### A Analogia dos Interruptores
Imagine que você tem $3$ peças ($A$, $B$ e $C$). Cada peça pode estar dentro ou fora da mochila. Isso é como ter 3 interruptores de luz (0 = desligado, 1 = ligado):

| Número Decimal | Em Binário (Interruptores C B A) | Peças Compradas |
| :---: | :---: | :--- |
| **0** | `0 0 0` | Nenhuma peça |
| **1** | `0 0 1` | Apenas a peça A |
| **2** | `0 1 0` | Apenas a peça B |
| **3** | `0 1 1` | Peças A e B |
| **4** | `1 0 0` | Apenas a peça C |
| **5** | `1 0 1` | Peças A e C |
| **6** | `1 1 0` | Peças B e C |
| **7** | `1 1 1` | Todas as peças (A, B e C) |

Note que um simples loop de `0` até `7` percorre **todas as combinações do universo**!

#### Como isso é feito em código Go:
```go
func ForcaBruta(pecas []Peca, orcamento int, kmAtual int) int {
    n := len(pecas)
    // ...
    melhor := 0

    // 1 << n é 2 elevado a n. Se n=3, vai de 0 até 7.
    for mascara := uint(0); mascara < (1 << uint(n)); mascara++ {
        custo, valor := 0, 0

        for i := 0; i < n; i++ {
            // Verifica se o interruptor da peça 'i' está ligado (bit 1)
            if mascara&(1<<uint(i)) != 0 {
                custo += pecas[i].Preco
                valor += valores[i]
            }
        }

        // Se couber no orçamento e for melhor que o recorde anterior:
        if custo <= orcamento && valor > melhor {
            melhor = valor
        }
    }
    return melhor
}
```

---

### O Gerador de Cenários (`gerarCenario`)

Para não testar sempre com os mesmos dados, a função `gerarCenario` simula motos e condições de trânsito variadas:
* Sorteia uma quilometragem atual entre **10.000 km e 50.000 km**.
* Gera entre 1 e 14 peças com preços aleatórios, vidas úteis diferentes e níveis de importância variados.
* Sorteia um orçamento surpresa.

---

### O Juiz Final: A Função `Validar`

A função `Validar(casos int)` roda um campeonato de 500 rodadas e faz 4 verificações rigorosas a cada rodada:

```go
func Validar(casos int) bool {
    // ...
    for c := 0; c < casos; c++ {
        // 1. Gera cenário aleatório
        pecas, orcamento, km := gerarCenario(rng, n)

        // 2. Roda os dois algoritmos
        escolhidas, valorDP := Mochila(pecas, orcamento, km)
        valorBF := ForcaBruta(pecas, orcamento, km)

        // TESTE 1: A pontuação da Mochila foi idêntica à da Força Bruta?
        if valorDP != valorBF {
            return false // FALHA: a mochila calculou errado!
        }

        // TESTE 2: Não comprou nenhuma peça repetida?
        // TESTE 3: O custo total respeitou o orçamento limite?
        // TESTE 4: A soma real dos itens bate com o total prometido pela matriz?
        if custo > orcamento || soma != valorDP {
            return false // FALHA: estourou orçamento ou soma incorreta!
        }
    }
    return true // PASSOU EM TUDO!
}
```

Se todas as 500 rodadas passarem sem nenhum erro, o console exibe com segurança:
> `OK: DP igual à força bruta em todos os cenários.`

---

## 5. Resumo das Diferenças

| Característica | Programação Dinâmica (`Mochila`) | Força Bruta (`ForcaBruta`) |
| :--- | :--- | :--- |
| **Como pensa?** | Inteligente: constrói a resposta reaproveitando cálculos anteriores. | Exaustiva: testa absolutamente todas as possibilidades. |
| **Velocidade** | Instantânea, mesmo com 100 peças e R$ 5.000 de orçamento. | Rápida para até 15 peças; impraticável para mais de 30 peças ($2^{30} > 1 \text{ bilhão}$ cálculos). |
| **Onde usamos?** | Na tela principal do usuário para entregar a resposta na hora. | Apenas no teste de validação para garantir que não há bugs. |

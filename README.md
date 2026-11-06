# 🏍️ Otimização de Troca de Peças (Problema da Mochila 0/1)

Algoritmo em **Go** para planejamento e otimização financeira na manutenção de motocicletas (baseado na *Bajaj Dominar NS 200*).

Utiliza a modelagem clássica do **Problema da Mochila 0/1 (*0/1 Knapsack Problem*) com Programação Dinâmica (DP)** para selecionar a combinação ideal de peças a serem trocadas com base no desgaste atual e na criticidade de cada item, respeitando rigorosamente o limite do orçamento.

---

## 🚀 Como Executar

Basta ter o [Go instalado](https://go.dev/) (versão 1.20+) e executar:

```bash
go run .
```

ou

```bash
go run main.go
```

> **Não é necessário passar nenhuma flag ou parâmetro!** O algoritmo executa diretamente com os dados de `pecas.csv`, avalia a quilometragem e o orçamento padrão (32.000 km e R$ 600,00) e valida automaticamente a solução contra força bruta em 500 cenários aleatórios.

---

## 📊 Estrutura do Arquivo `pecas.csv`

O CSV contém os dados das peças da moto:

| Coluna | Descrição |
| :--- | :--- |
| `nome` | Nome da peça/componente |
| `preco` | Custo estimado de troca/compra (R$) |
| `km_usada` | Quilometragem percorrida na última troca |
| `vida_util_km` | Vida útil recomendada pelo manual (km) |
| `importancia` | Grau de criticidade/segurança (1 a 10) |

---

## 🧠 Como Funciona o Algoritmo

1. **Cálculo de Desgaste e Urgência:**
   - Desgaste relativo: `(kmAtual - kmUltimaTroca) / vidaUtil` (limitado a 100%).
   - Urgência (valor): `desgaste * criticidade * 100`, com bônus de risco se a peça atingiu ou superou 100% de desgaste.
2. **Programação Dinâmica:**
   - Constrói uma matriz de decisão `dp[i][w]`, onde cada estado reaproveita o melhor resultado obtido para orçamentos menores.
3. **Reconstrução das Peças:**
   - Percorre a matriz de trás para frente para listar exatamente quais peças foram selecionadas e quais ficaram de fora para o próximo ciclo.
4. **Validação com Oráculo (Força Bruta):**
   - Compara a resposta da Programação Dinâmica contra todas as $2^n$ combinações possíveis em 500 testes gerados aleatoriamente.

Para uma explicação detalhada passo a passo, confira o arquivo [EXPLICACAO.md](EXPLICACAO.md).

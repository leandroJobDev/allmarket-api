# ⚙️ AllM@rket - API

> **Motor de processamento distribuído e parsing de dados fiscais em tempo real.**

Este repositório contém a API robusta do AllM@rket, desenvolvida em **Go (Golang)**. O sistema é responsável pela orquestração de dados, web scraping de alta performance e integração com bases de dados NoSQL, servindo como o *backbone* para o ecossistema de Micro-frontends.

## 🏗️ Arquitetura de Software

A API foi desenhada seguindo os princípios de **Clean Architecture**, garantindo que a lógica de negócio seja independente de frameworks e drivers externos:

* **Entities:** Modelagem de dados central (Notas, Itens, Estabelecimentos).
* **Use Cases:** Orquestração de regras de negócio (Processamento de QR Code, Validação de acesso).
* **Infrastructure:** Implementações técnicas (Firestore, Gin Router, Parsers SEFAZ).
* **Repository Pattern:** Abstração da camada de persistência para facilitar testes e escalabilidade.

---

## 💡 A Filosofia dos 3 Es (Backend Edition)

No backend, os pilares do sistema são traduzidos em eficiência computacional:

* **Escaneia (Parsing Engine):** Utilização de `goquery` para realizar o *scraping* e *parsing* de HTML/XML da SEFAZ em milissegundos, convertendo cupons complexos em estruturas JSON otimizadas.
* **Examina (Data Intelligence):** Algoritmos de normalização de dados que identificam itens e estabelecimentos em diferentes estados (SP, SC, PE, PB), tratando inconsistências das notas fiscais.
* **Economiza (API Performance):** O uso de Go garante um consumo mínimo de memória e CPU, permitindo que a inteligência de comparação de preços seja entregue com baixíssima latência.

---

## 🛠️ Stack Tecnológica

* **Linguagem:** Go 1.25 (Alta performance e concorrência).
* **Framework HTTP:** `Gin Gonic` (Roteamento de ultra-velocidade).
* **Database:** `Google Cloud Firestore` (Persistência NoSQL escalável).
* **Extraction:** `goquery` (Web scraping e parsing de documentos fiscais).
* **Config:** `godotenv` (Gestão de ambiente segura).
* **Containerização:** Docker com *Multi-stage Build* (Imagem final baseada em Alpine Linux com < 20MB).

---

## 🚀 Engenharia de Operações

### Endpoints Principais (REST)

* `POST /api/v1/notas`: Processa a URL do QR Code e extrai os itens.
* `GET /api/v1/notas/:email`: Recupera o histórico de consumo por usuário.
* `DELETE /api/v1/notas/:id`: Gestão de registros.

### DevOps & CI/CD

A API foi preparada para ambientes Cloud Native:

```bash
# Build da imagem Docker otimizada
docker build -t allmarket-api .

# Execução local via Docker
docker run -p 8080:8080 --env-file .env allmarket-api

```

---

## ✨ Diferenciais Técnicos

1. **Concorrência Nativa:** Aproveitamento de *Goroutines* para processos de extração paralela.
2. **Parsing Multi-Estado:** Engine adaptativa capaz de ler formatos distintos de SEFAZs estaduais (HTML e XML).
3. **Segurança:** Implementação de política rigorosa de CORS e sanitização de entradas para proteção contra injeções.
4. **Testabilidade:** Cobertura de testes unitários nas camadas de Use Case e Entity, garantindo a integridade do motor de cálculo.

---


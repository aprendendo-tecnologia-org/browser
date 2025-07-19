# browser

O pacote `browser` fornece uma abstração de alto nível para automação de navegação web utilizando o [chromedp](https://github.com/chromedp/chromedp) em Go. Ele permite visitar páginas, extrair documentos HTML, interagir com elementos do DOM, aguardar elementos específicos, clicar em elementos e executar funções personalizadas sobre nós do DOM.

## Propósito

O objetivo deste pacote é facilitar a automação de tarefas em navegadores, como scraping, testes automatizados de interfaces web e manipulação de páginas dinâmicas, sem a necessidade de lidar diretamente com a API detalhada do chromedp.

## Principais Funcionalidades

- **Visitar URLs**: Navega até uma URL e armazena o documento HTML retornado.
- **Obter documento HTML**: Recupera o HTML da página atual.
- **Obter URL atual**: Retorna a URL da página atualmente carregada.
- **Aguardar elementos**: Espera até que um seletor CSS esteja visível na página.
- **Obter nós do DOM**: Extrai nós do DOM que correspondem a um seletor.
- **Clicar em elementos**: Realiza cliques em elementos da página.
- **Executar funções customizadas**: Permite executar funções sobre nós do DOM encontrados.
- **Configurar timeout**: Permite definir o tempo máximo de espera para operações do navegador.

## Sobre o Status Code

O método `Visit` retorna o código de status HTTP (`StatusCode`) da resposta ao navegar para a URL desejada.  
**A decisão sobre como tratar códigos de status diferentes de 200 é responsabilidade do usuário da API.**  
O pacote não gera erro automaticamente para códigos de status diferentes de 200, permitindo flexibilidade para diferentes cenários de uso, como scraping de páginas 404 ou análise de respostas específicas.

---

## Exemplos de Uso

### 1. Visitando uma página e obtendo o HTML

```go
ctx, cancel := chromedp.NewContext(context.Background())
defer cancel()

b := browser.NewBrowser()
status, err := b.Visit(ctx, "https://exemplo.com")
if err != nil {
    log.Fatalf("Erro ao visitar página: %v", err)
}
fmt.Printf("Status: %d\n", status)
fmt.Println("HTML:", b.GetDocument())
```

### 2. Usando um servidor de teste para fornecer HTML

```go
const html = `
    <html><body>
        <div id="unique">Hello Node</div>
        <ul>
            <li class="product">Produto A</li>
            <li class="product">Produto B</li>
        </ul>
    </body></html>
`
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}))
defer server.Close()

b := browser.NewBrowser()
b.Visit(ctx, server.URL)
nodes, err := b.GetNodes(ctx, ".product")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Encontrados %d produtos\n", len(nodes))
```

### 3. Interagindo com elementos e aguardando alterações no DOM

```go
const html = `
    <html><body>
        <button id="createBtn" onclick="setTimeout(function() {
            var el = document.createElement('span');
            el.id = 'created';
            el.textContent = 'Elemento criado!';
            document.getElementById('container').appendChild(el);
        }, 1200);">Criar elemento</button>
        <div id="container"></div>
    </body></html>
`
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}))
defer server.Close()

b := browser.NewBrowser()
b.Visit(ctx, server.URL)
b.Click(ctx, "#createBtn")
b.WaitForElement(ctx, "#created")
doc := b.GetDocument()
fmt.Println("Documento após clique:", doc)
```

### 4. Extraindo links externos de uma página

```go
const html = `
    <html><body>
        <a href="https://site1.com" class="external">Site 1</a>
        <a href="https://site2.com" class="external">Site 2</a>
        <a href="/internal" class="internal">Internal</a>
    </body></html>
`
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}))
defer server.Close()

b := browser.NewBrowser()
b.Visit(ctx, server.URL)
var foundLinks []string
b.Exec(ctx, "a.external", func(nodes []*cdp.Node) {
    for _, node := range nodes {
        for i := 0; i < len(node.Attributes); i += 2 {
            if node.Attributes[i] == "href" {
                foundLinks = append(foundLinks, node.Attributes[i+1])
            }
        }
    }
})
fmt.Printf("Links externos encontrados: %v\n", foundLinks)
```

### 5. Exemplo Prático de Uso do Pacote

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/chromedp/chromedp"
    "com.github/aprendendo-tecnologia-org/browser"
)

func main() {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    b := browser.NewBrowser()
    status, err := b.Visit(ctx, "https://exemplo.com/produtos")
    if err != nil {
        log.Fatalf("Erro ao visitar página: %v", err)
    }
    if status != 200 {
        log.Fatalf("Status HTTP inesperado: %d", status)
    }

    // Aguarda a lista de produtos aparecer
    if err := b.WaitForElement(ctx, ".product"); err != nil {
        log.Fatalf("Produto não encontrado: %v", err)
    }

    // Extrai nomes dos produtos
    var produtos []string
    err = b.Exec(ctx, ".product", func(nodes []*cdp.Node) {
        for _, node := range nodes {
            if node.NodeValue != "" {
                produtos = append(produtos, node.NodeValue)
            } else {
                // Busca em filhos se necessário
                for _, child := range node.Children {
                    if child.NodeValue != "" {
                        produtos = append(produtos, child.NodeValue)
                    }
                }
            }
        }
    })
    if err != nil {
        log.Fatalf("Erro ao extrair produtos: %v", err)
    }

    fmt.Println("Produtos encontrados:")
    for _, nome := range produtos {
        fmt.Println("-", nome)
    }

    // Exemplo: clicar em um botão para carregar mais produtos
    if err := b.Click(ctx, "#loadMore"); err == nil {
        b.WaitForElement(ctx, ".product.new")
        // ... repetir extração se necessário
    }
}
```

## Instalação

Para instalar o pacote, utilize o comando:

```sh
go get com.github/aprendendo-tecnologia-org/browser
```

Certifique-se também de instalar o [chromedp](https://github.com/chromedp/chromedp) como dependência, caso ainda não esteja presente no seu projeto:

```sh
go get github.com/chromedp/chromedp
```

Inclua o import no seu código Go:

```go
import "com.github/aprendendo-tecnologia-org/browser"
```